package db

import (
	"fmt"
	"log"

	"github.com/lukaszbudnik/migrator/config"
)

// Dialect returns SQL statements for given DB
type Dialect interface {
	GetTenantInsertSQL() string
	GetTenantSelectSQL() string
	GetMigrationInsertSQL() string
	GetMigrationSelectSQL() string
	GetCreateTenantsTableSQL() string
	GetCreateMigrationsTableSQL() string
	GetCreateSchemaSQL(string) string
}

// BaseDialect struct is used to provide default Dialect interface implementation
type BaseDialect struct {
}

const (
	selectMigrationsSQL      = "select name, source_dir as sd, filename, type, db_schema, created, contents, checksum from %v.%v order by name, source_dir"
	selectTenantsSQL         = "select name from %v.%v"
	createMigrationsTableSQL = `
create table if not exists %v.%v (
  id serial primary key,
  name varchar(200) not null,
  source_dir varchar(200) not null,
  filename varchar(200) not null,
  type int not null,
  db_schema varchar(200) not null,
  created timestamp default now(),
	contents text,
	checksum varchar(64)
)
`
	createTenantsTableSQL = `
create table if not exists %v.%v (
  id serial primary key,
  name varchar(200) not null,
  created timestamp default now()
)
`
	createSchemaSQL = "create schema if not exists %v"
)

// GetCreateTenantsTableSQL returns migrator's default create tenants table SQL statement.
// This SQL is used by both MySQL and PostgreSQL.
func (bd *BaseDialect) GetCreateTenantsTableSQL() string {
	return fmt.Sprintf(createTenantsTableSQL, migratorSchema, migratorTenantsTable)
}

// GetCreateMigrationsTableSQL returns migrator's create migrations table SQL statement.
// This SQL is used by both MySQL and PostgreSQL.
func (bd *BaseDialect) GetCreateMigrationsTableSQL() string {
	return fmt.Sprintf(createMigrationsTableSQL, migratorSchema, migratorMigrationsTable)
}

// GetTenantSelectSQL returns migrator's default tenant select SQL statement.
// This SQL is used by all MySQL, PostgreSQL, and MS SQL.
func (bd *BaseDialect) GetTenantSelectSQL() string {
	return fmt.Sprintf(selectTenantsSQL, migratorSchema, migratorTenantsTable)
}

// GetMigrationSelectSQL returns migrator's migrations select SQL statement.
// This SQL is used by all MySQL, PostgreSQL, MS SQL.
func (bd *BaseDialect) GetMigrationSelectSQL() string {
	return fmt.Sprintf(selectMigrationsSQL, migratorSchema, migratorMigrationsTable)
}

// GetCreateSchemaSQL returns create schema SQL statement.
// This SQL is used by both MySQL and PostgreSQL.
func (bd *BaseDialect) GetCreateSchemaSQL(schema string) string {
	return fmt.Sprintf(createSchemaSQL, schema)
}

// CreateDialect constructs Dialect instance based on the passed Config
func CreateDialect(config *config.Config) Dialect {

	var dialect Dialect

	switch config.Driver {
	case "mysql":
		dialect = &mySQLDialect{}
	case "sqlserver":
		dialect = &msSQLDialect{}
	case "postgres":
		dialect = &postgreSQLDialect{}
	default:
		log.Panicf("Failed to create Connector: %q is an unknown driver.", config.Driver)
	}

	return dialect
}
