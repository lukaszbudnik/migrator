package db

import (
	"database/sql"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"log"
	"strings"
	"time"
)

// Connector interface abstracts all DB operations performed by migrator
type Connector interface {
	Init()
	GetTenantSelectSQL() string
	GetTenantInsertSQL() string
	AddTenantAndApplyMigrations(string, []types.Migration)
	GetTenants() []string
	GetSchemaPlaceHolder() string
	GetDBMigrations() []types.MigrationDB
	ApplyMigrations(migrations []types.Migration)
	Dispose()
}

// BaseConnector struct is a base struct for implementing DB specific dialects
type BaseConnector struct {
	Config  *config.Config
	Dialect Dialect
	DB      *sql.DB
}

// CreateConnector constructs Connector instance based on the passed Config
func CreateConnector(config *config.Config) Connector {
	dialect := CreateDialect(config)
	connector := &BaseConnector{config, dialect, nil}
	return connector
}

const (
	migratorSchema           = "migrator"
	migratorTenantsTable     = "migrator_tenants"
	migratorMigrationsTable  = "migrator_migrations"
	defaultSchemaPlaceHolder = "{schema}"
)

// Init initialises connector by opening a connection to database
func (bc *BaseConnector) Init() {
	db, err := sql.Open(bc.Config.Driver, bc.Config.DataSource)
	if err != nil {
		log.Panicf("Failed to create database connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Panicf("Failed to connect to database: %v", err)
	}
	bc.DB = db

	tx, err := bc.DB.Begin()
	if err != nil {
		log.Panicf("Could not start DB transaction: %v", err)
	}

	// make sure migrator schema exists
	createSchema := bc.Dialect.GetCreateSchemaSQL(migratorSchema)
	if _, err := bc.DB.Query(createSchema); err != nil {
		log.Panicf("Could not create migrator schema: %v", err)
	}

	// make sure migrations table exists
	createMigrationsTable := bc.Dialect.GetCreateMigrationsTableSQL()
	if _, err := bc.DB.Query(createMigrationsTable); err != nil {
		log.Panicf("Could not create migrations table: %v", err)
	}

	// if using default migrator tenants table make sure it exists
	if bc.Config.TenantSelectSQL == "" {
		createTenantsTable := bc.Dialect.GetCreateTenantsTableSQL()
		if _, err := bc.DB.Query(createTenantsTable); err != nil {
			log.Panicf("Could not create default tenants table: %v", err)
		}
	}

	tx.Commit()
}

// Dispose closes all resources allocated by connector
func (bc *BaseConnector) Dispose() {
	if bc.DB != nil {
		bc.DB.Close()
	}
}

// GetTenantSelectSQL returns SQL to be executed to list all DB tenants
func (bc *BaseConnector) GetTenantSelectSQL() string {
	var tenantSelectSQL string
	if bc.Config.TenantSelectSQL != "" {
		tenantSelectSQL = bc.Config.TenantSelectSQL
	} else {
		tenantSelectSQL = bc.Dialect.GetTenantSelectSQL()
	}
	return tenantSelectSQL
}

// GetTenants returns a list of all DB tenants
func (bc *BaseConnector) GetTenants() []string {
	tenantSelectSQL := bc.GetTenantSelectSQL()

	rows, err := bc.DB.Query(tenantSelectSQL)
	if err != nil {
		log.Panicf("Could not query tenants: %v", err)
	}
	var tenants []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Panicf("Could not read tenants: %v", err)
		}
		tenants = append(tenants, name)
	}
	return tenants
}

// GetDBMigrations returns a list of all applied DB migrations
func (bc *BaseConnector) GetDBMigrations() []types.MigrationDB {
	query := bc.Dialect.GetMigrationSelectSQL()

	rows, err := bc.DB.Query(query)
	if err != nil {
		log.Panicf("Could not query DB migrations: %v", err)
	}

	var dbMigrations []types.MigrationDB
	for rows.Next() {
		var (
			name          string
			sourceDir     string
			filename      string
			migrationType types.MigrationType
			schema        string
			created       time.Time
		)
		if err := rows.Scan(&name, &sourceDir, &filename, &migrationType, &schema, &created); err != nil {
			log.Panicf("Could not read DB migration: %v", err)
		}
		mdef := types.MigrationDefinition{Name: name, SourceDir: sourceDir, File: filename, MigrationType: migrationType}
		dbMigrations = append(dbMigrations, types.MigrationDB{MigrationDefinition: mdef, Schema: schema, Created: created})
	}

	return dbMigrations
}

// ApplyMigrations applies passed migrations
func (bc *BaseConnector) ApplyMigrations(migrations []types.Migration) {
	if len(migrations) == 0 {
		return
	}

	tenants := bc.GetTenants()

	tx, err := bc.DB.Begin()
	if err != nil {
		log.Panicf("Could not start DB transaction: %v", err)
	}

	bc.applyMigrationsInTx(tx, tenants, migrations)

	tx.Commit()
}

// AddTenantAndApplyMigrations adds new tenant and applies all existing tenant migrations
func (bc *BaseConnector) AddTenantAndApplyMigrations(tenant string, migrations []types.Migration) {
	tenantInsertSQL := bc.GetTenantInsertSQL()

	tx, err := bc.DB.Begin()
	if err != nil {
		log.Panicf("Could not start DB transaction: %v", err)
	}

	createSchema := bc.Dialect.GetCreateSchemaSQL(tenant)
	if _, err = tx.Exec(createSchema); err != nil {
		tx.Rollback()
		log.Panicf("Create schema failed, transaction rollback was called: %v", err)
	}

	insert, err := bc.DB.Prepare(tenantInsertSQL)
	if err != nil {
		log.Panicf("Could not create prepared statement: %v", err)
	}

	_, err = tx.Stmt(insert).Exec(tenant)
	if err != nil {
		tx.Rollback()
		log.Panicf("Failed to add tenant entry, transaction rollback was called: %v", err)
	}

	bc.applyMigrationsInTx(tx, []string{tenant}, migrations)

	tx.Commit()
}

// GetTenantInsertSQL returns tenant insert SQL statement from configuration file
// or, if absent, returns default Dialect-specific migrator tenant insert SQL
func (bc *BaseConnector) GetTenantInsertSQL() string {
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
// either the default one or overriden by user in config
func (bc *BaseConnector) GetSchemaPlaceHolder() string {
	var schemaPlaceHolder string
	if bc.Config.SchemaPlaceHolder != "" {
		schemaPlaceHolder = bc.Config.SchemaPlaceHolder
	} else {
		schemaPlaceHolder = defaultSchemaPlaceHolder
	}
	return schemaPlaceHolder
}

func (bc *BaseConnector) applyMigrationsInTx(tx *sql.Tx, tenants []string, migrations []types.Migration) {
	schemaPlaceHolder := bc.GetSchemaPlaceHolder()

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
				tx.Rollback()
				log.Panicf("SQL failed, transaction rollback was called: %v %v", err, contents)
			}

			_, err = tx.Stmt(insert).Exec(m.Name, m.SourceDir, m.File, m.MigrationType, s)
			if err != nil {
				tx.Rollback()
				log.Panicf("Failed to add migration entry, transaction rollback was called: %v", err)
			}
		}

	}
}
