package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

// Connector interface abstracts all DB operations performed by migrator
type Connector interface {
	Init() error
	GetTenants() ([]string, error)
	GetDBMigrations() ([]types.MigrationDB, error)
	ApplyMigrations(migrations []types.Migration) error
	AddTenantAndApplyMigrations(string, []types.Migration) error
	Dispose()
}

// BaseConnector struct is a base struct for implementing DB specific dialects
type BaseConnector struct {
	Config  *config.Config
	Dialect dialect
	DB      *sql.DB
}

// NewConnector constructs Connector instance based on the passed Config
func NewConnector(config *config.Config) (Connector, error) {
	dialect, err := newDialect(config)
	if err != nil {
		return nil, err
	}
	connector := &BaseConnector{config, dialect, nil}
	return connector, nil
}

const (
	migratorSchema           = "migrator"
	migratorTenantsTable     = "migrator_tenants"
	migratorMigrationsTable  = "migrator_migrations"
	defaultSchemaPlaceHolder = "{schema}"
)

// Init initialises connector by opening a connection to database
func (bc *BaseConnector) Init() error {
	db, err := sql.Open(bc.Config.Driver, bc.Config.DataSource)
	if err != nil {
		return err
	}
	return bc.doInit(db)
}

// Init initialises connector by opening a connection to database
func (bc *BaseConnector) doInit(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("Failed to connect to database: %v", err)
	}
	bc.DB = db

	tx, err := bc.DB.Begin()
	if err != nil {
		return fmt.Errorf("Could not start DB transaction: %v", err)
	}

	// make sure migrator schema exists
	createSchema := bc.Dialect.GetCreateSchemaSQL(migratorSchema)
	if _, err := bc.DB.Query(createSchema); err != nil {
		return fmt.Errorf("Could not create migrator schema: %v", err)
	}

	// make sure migrations table exists
	createMigrationsTable := bc.Dialect.GetCreateMigrationsTableSQL()
	if _, err := bc.DB.Query(createMigrationsTable); err != nil {
		return fmt.Errorf("Could not create migrations table: %v", err)
	}

	// if using default migrator tenants table make sure it exists
	if bc.Config.TenantSelectSQL == "" {
		createTenantsTable := bc.Dialect.GetCreateTenantsTableSQL()
		if _, err := bc.DB.Query(createTenantsTable); err != nil {
			return fmt.Errorf("Could not create default tenants table: %v", err)
		}
	}

	return tx.Commit()
}

// Dispose closes all resources allocated by connector
func (bc *BaseConnector) Dispose() {
	if bc.DB != nil {
		bc.DB.Close()
	}
}

// getTenantSelectSQL returns SQL to be executed to list all DB tenants
func (bc *BaseConnector) getTenantSelectSQL() string {
	var tenantSelectSQL string
	if bc.Config.TenantSelectSQL != "" {
		tenantSelectSQL = bc.Config.TenantSelectSQL
	} else {
		tenantSelectSQL = bc.Dialect.GetTenantSelectSQL()
	}
	return tenantSelectSQL
}

// GetTenants returns a list of all DB tenants
func (bc *BaseConnector) GetTenants() (tenants []string, err error) {
	tenantSelectSQL := bc.getTenantSelectSQL()

	rows, err := bc.DB.Query(tenantSelectSQL)
	if err != nil {
		err = fmt.Errorf("Could not query tenants: %v", err)
		return
	}

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			err = fmt.Errorf("Could not read tenants: %v", err)
			return
		}
		tenants = append(tenants, name)
	}
	return
}

// GetDBMigrations returns a list of all applied DB migrations
func (bc *BaseConnector) GetDBMigrations() (dbMigrations []types.MigrationDB, err error) {
	query := bc.Dialect.GetMigrationSelectSQL()

	rows, err := bc.DB.Query(query)
	if err != nil {
		err = fmt.Errorf("Could not query DB migrations: %v", err)
		return
	}

	for rows.Next() {
		var (
			name          string
			sourceDir     string
			filename      string
			migrationType types.MigrationType
			schema        string
			created       time.Time
			contents      string
			checksum      string
		)
		if err = rows.Scan(&name, &sourceDir, &filename, &migrationType, &schema, &created, &contents, &checksum); err != nil {
			err = fmt.Errorf("Could not read DB migration: %v", err)
			return
		}
		mdef := types.Migration{Name: name, SourceDir: sourceDir, File: filename, MigrationType: migrationType, Contents: contents, CheckSum: checksum}
		dbMigrations = append(dbMigrations, types.MigrationDB{Migration: mdef, Schema: schema, Created: created})
	}

	return
}

// ApplyMigrations applies passed migrations
func (bc *BaseConnector) ApplyMigrations(migrations []types.Migration) (err error) {
	if len(migrations) == 0 {
		return
	}

	tenants, err := bc.GetTenants()
	if err != nil {
		return
	}

	tx, err := bc.DB.Begin()
	if err != nil {
		return
	}

	defer func() {
		r := recover()
		if r == nil {
			err = tx.Commit()
		} else {
			log.Println("Recovered in ApplyMigrations. Transaction rollback.")
			tx.Rollback()
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	bc.applyMigrationsInTx(tx, tenants, migrations)
	return
}

// AddTenantAndApplyMigrations adds new tenant and applies all existing tenant migrations
func (bc *BaseConnector) AddTenantAndApplyMigrations(tenant string, migrations []types.Migration) (err error) {
	tenantInsertSQL := bc.getTenantInsertSQL()

	tx, err := bc.DB.Begin()
	if err != nil {
		return
	}

	defer func() {
		r := recover()
		if r == nil {
			err = tx.Commit()
		} else {
			log.Println("Recovered in AddTenantAndApplyMigrations. Transaction rollback.")
			tx.Rollback()
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	createSchema := bc.Dialect.GetCreateSchemaSQL(tenant)
	if _, err = tx.Exec(createSchema); err != nil {
		log.Panicf("Create schema failed, transaction rollback was called: %v", err)
	}

	insert, err := bc.DB.Prepare(tenantInsertSQL)
	if err != nil {
		log.Panicf("Could not create prepared statement: %v", err)
	}

	_, err = tx.Stmt(insert).Exec(tenant)
	if err != nil {
		log.Panicf("Failed to add tenant entry: %v", err)
	}

	bc.applyMigrationsInTx(tx, []string{tenant}, migrations)

	return
}

// getTenantInsertSQL returns tenant insert SQL statement from configuration file
// or, if absent, returns default Dialect-specific migrator tenant insert SQL
func (bc *BaseConnector) getTenantInsertSQL() string {
	var tenantInsertSQL string
	// if set explicitly in config use it
	// otherwise use default value provided by Dialect implementation
	if bc.Config.TenantInsertSQL != "" {
		tenantInsertSQL = bc.Config.TenantInsertSQL
	} else {
		tenantInsertSQL = bc.Dialect.GetTenantInsertSQL()
	}
	return tenantInsertSQL
}

// GetSchemaPlaceHolder returns a schema placeholder which is
// either the default one or overridden by user in config
func (bc *BaseConnector) getSchemaPlaceHolder() string {
	var schemaPlaceHolder string
	if bc.Config.SchemaPlaceHolder != "" {
		schemaPlaceHolder = bc.Config.SchemaPlaceHolder
	} else {
		schemaPlaceHolder = defaultSchemaPlaceHolder
	}
	return schemaPlaceHolder
}

func (bc *BaseConnector) applyMigrationsInTx(tx *sql.Tx, tenants []string, migrations []types.Migration) {
	schemaPlaceHolder := bc.getSchemaPlaceHolder()

	insertMigrationSQL := bc.Dialect.GetMigrationInsertSQL()
	insert, err := bc.DB.Prepare(insertMigrationSQL)
	if err != nil {
		log.Panicf("Could not create prepared statement: %v", err)
	}

	for _, m := range migrations {
		var schemas []string
		if m.MigrationType == types.MigrationTypeTenantSchema {
			schemas = tenants
		} else {
			schemas = []string{m.SourceDir}
		}

		for _, s := range schemas {
			log.Printf("Applying migration type: %d, schema: %s, file: %s ", m.MigrationType, s, m.File)

			contents := strings.Replace(m.Contents, schemaPlaceHolder, s, -1)

			_, err = tx.Exec(contents)
			if err != nil {
				log.Panicf("SQL migration failed: %v", err)
			}

			_, err = tx.Stmt(insert).Exec(m.Name, m.SourceDir, m.File, m.MigrationType, s, m.Contents, m.CheckSum)
			if err != nil {
				log.Panicf("Failed to add migration entry: %v", err)
			}
		}

	}
}
