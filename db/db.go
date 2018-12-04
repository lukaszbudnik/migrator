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
	GetTenantSelectSql() string
	GetTenantInsertSql() string
	AddTenantAndApplyMigrations(string, []types.Migration)
	GetTenants() []string
	GetSchemaPlaceHolder() string
	// rename to GetDBMigrations, change type to DBMigration
	GetMigrations() []types.MigrationDB
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

	// make sure migrations table exists
	createMigrationsTable := bc.Dialect.GetCreateMigrationsTableSql()
	if _, err := bc.DB.Query(createMigrationsTable); err != nil {
		log.Panicf("Could not create migrations table: %v", err)
	}

	// if using default migrator tenants table make sure it exists
	if bc.Config.TenantSelectSql == "" {
		createTenantsTable := bc.Dialect.GetCreateTenantsTableSql()
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

// GettenantsSql returns SQL to be executed to list all DB tenants
func (bc *BaseConnector) GetTenantSelectSql() string {
	var tenantSelectSql string
	if bc.Config.TenantSelectSql != "" {
		tenantSelectSql = bc.Config.TenantSelectSql
	} else {
		tenantSelectSql = bc.Dialect.GetTenantSelectSql()
	}
	return tenantSelectSql
}

// GetTenants returns a list of all DB tenants as specified by
// defaultSelectTenants or the value specified in config
func (bc *BaseConnector) GetTenants() []string {
	tenantSelectSql := bc.GetTenantSelectSql()

	rows, err := bc.DB.Query(tenantSelectSql)
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

// GetMigrations returns a list of all applied DB migrations
func (bc *BaseConnector) GetMigrations() []types.MigrationDB {
	query := bc.Dialect.GetMigrationSelectSql()

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
		mdef := types.MigrationDefinition{name, sourceDir, filename, migrationType}
		dbMigrations = append(dbMigrations, types.MigrationDB{mdef, schema, created})
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
	tenantInsertSql := bc.GetTenantInsertSql()

	tx, err := bc.DB.Begin()
	if err != nil {
		log.Panicf("Could not start DB transaction: %v", err)
	}

	insert, err := bc.DB.Prepare(tenantInsertSql)
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

func (bc *BaseConnector) GetTenantInsertSql() string {
	var tenantInsertSql string
	// if set explicitly in config use it
	// otherwise use default value provided by Dialect implementation
	if bc.Config.TenantInsertSql != "" {
		tenantInsertSql = bc.Config.TenantInsertSql
	} else {
		tenantInsertSql = bc.Dialect.GetTenantInsertSql()
	}
	return tenantInsertSql
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

	insertMigrationSql := bc.Dialect.GetMigrationInsertSql()

	insert, err := bc.DB.Prepare(insertMigrationSql)
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
