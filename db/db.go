package db

import (
	"database/sql"
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"log"
	"strings"
	"time"
)

// Connector interface abstracts all DB operations performed by migrator
type Connector interface {
	Init()
	GetTenantsSQL() string
	GetTenants() []string
	GetSchemaPlaceHolder() string
	GetMigrations() []types.MigrationDB
	ApplyMigrations(migrations []types.Migration)
	Dispose()
}

// BaseConnector struct is a base struct for implementing DB specific dialects
type BaseConnector struct {
	Config *config.Config
	DB     *sql.DB
}

// CreateConnector constructs Connector instance based on the passed Config
func CreateConnector(config *config.Config) Connector {
	var bc = BaseConnector{config, nil}
	var connector Connector

	switch config.Driver {
	case "mysql":
		connector = &mySQLConnector{bc}
	case "postgres":
		connector = &postgreSQLConnector{bc}
	default:
		log.Panicf("Failed to create Connector: %q is an unknown driver.", config.Driver)
	}

	return connector
}

const (
	migrationsTableName     = "public.migrator_migrations"
	defaultTenantsTableName = "public.migrator_tenants"
	createMigrationsTable   = `
  create table if not exists %v (
    id serial primary key,
    name varchar(200) not null,
		source_dir varchar(200) not null,
    file varchar(200) not null,
		type int not null,
    db_schema varchar(200) not null,
    created timestamp default now()
	)
  `
	createDefaultTenantsTable = `
	create table if not exists %v (
		id serial primary key,
		name varchar(200) not null,
		created timestamp default now()
	)
	`
	selectMigrations         = "select name, source_dir, file, type, db_schema, created from %v order by name, source_dir"
	defaultTenantsSQLPattern = "select name from %v"
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

	defaultTenantsSQL := fmt.Sprintf(defaultTenantsSQLPattern, defaultTenantsTableName)
	if bc.Config.TenantsSQL != "" && bc.Config.TenantsSQL != defaultTenantsSQL {
		createTableQuery := fmt.Sprintf(createDefaultTenantsTable, defaultTenantsTableName)

		if _, err := bc.DB.Query(createTableQuery); err != nil {
			log.Panicf("Could not create default tenants table: %v", err)
		}
	}

}

// Dispose closes all resources allocated by connector
func (bc *BaseConnector) Dispose() {
	if bc.DB != nil {
		bc.DB.Close()
	}
}

// GetTenantsSQL returns SQL to be execute to list all DB tenants
func (bc *BaseConnector) GetTenantsSQL() string {
	var tenantsSQL string
	if bc.Config.TenantsSQL != "" {
		tenantsSQL = bc.Config.TenantsSQL
	} else {
		tenantsSQL = fmt.Sprintf(defaultTenantsSQLPattern, defaultTenantsTableName)
	}
	return tenantsSQL
}

// GetTenants returns a list of all DB tenants as specified by
// defaultSelectTenants or the value specified in config
func (bc *BaseConnector) GetTenants() []string {
	tenantsSQL := bc.GetTenantsSQL()

	rows, err := bc.DB.Query(tenantsSQL)
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
	createTableQuery := fmt.Sprintf(createMigrationsTable, migrationsTableName)
	if _, err := bc.DB.Query(createTableQuery); err != nil {
		log.Panicf("Could not create migrations table: %v", err)
	}

	query := fmt.Sprintf(selectMigrations, migrationsTableName)

	rows, err := bc.DB.Query(query)
	if err != nil {
		log.Panicf("Could not query DB migrations: %v", err)
	}

	var dbMigrations []types.MigrationDB
	for rows.Next() {
		var (
			name          string
			sourceDir     string
			file          string
			migrationType types.MigrationType
			schema        string
			created       time.Time
		)
		if err := rows.Scan(&name, &sourceDir, &file, &migrationType, &schema, &created); err != nil {
			log.Panicf("Could not read DB migration: %v", err)
		}
		mdef := types.MigrationDefinition{name, sourceDir, file, migrationType}
		dbMigrations = append(dbMigrations, types.MigrationDB{mdef, schema, created})
	}

	return dbMigrations
}

// ApplyMigrations applies passed migrations
func (bc *BaseConnector) ApplyMigrations(migrations []types.Migration) {
	log.Panic("ApplyMigrations() must be overwritten by specific connector")
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

// applyMigrationsWithInsertMigrationSQL is called by specific implementations
// insertMigrationSQL varies based on database dialect
func (bc *BaseConnector) applyMigrationsWithInsertMigrationSQL(migrations []types.Migration, insertMigrationSQL string) {

	if len(migrations) == 0 {
		return
	}

	schemaPlaceHolder := bc.GetSchemaPlaceHolder()

	tenants := bc.GetTenants()

	tx, err := bc.DB.Begin()
	if err != nil {
		log.Panicf("Could not start DB transaction: %v", err)
	}

	query := fmt.Sprintf(insertMigrationSQL, migrationsTableName)
	insert, err := bc.DB.Prepare(query)
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
				log.Panicf("SQL failed, transaction rollback was called: %v", err)
			}

			_, err = tx.Stmt(insert).Exec(m.Name, m.SourceDir, m.File, m.MigrationType, s)
			if err != nil {
				tx.Rollback()
				log.Panicf("Failed to add migration entry, transaction rollback was called: %v", err)
			}
		}

	}

	tx.Commit()
}
