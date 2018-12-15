package db

import (
	"fmt"

	"github.com/lukaszbudnik/migrator/config"
)

// dialect returns SQL statements for given DB
type dialect interface {
	GetTenantInsertSQL() string
	GetTenantSelectSQL() string
	GetMigrationInsertSQL() string
	GetMigrationSelectSQL() string
	GetCreateTenantsTableSQL() string
	GetCreateMigrationsTableSQL() string
	GetCreateSchemaSQL(string) string
}

// baseDialect struct is used to provide default dialect interface implementation
type baseDialect struct {
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
func (bd *baseDialect) GetCreateTenantsTableSQL() string {
	return fmt.Sprintf(createTenantsTableSQL, migratorSchema, migratorTenantsTable)
}

// GetCreateMigrationsTableSQL returns migrator's create migrations table SQL statement.
// This SQL is used by both MySQL and PostgreSQL.
func (bd *baseDialect) GetCreateMigrationsTableSQL() string {
	return fmt.Sprintf(createMigrationsTableSQL, migratorSchema, migratorMigrationsTable)
}

// GetTenantSelectSQL returns migrator's default tenant select SQL statement.
// This SQL is used by all MySQL, PostgreSQL, and MS SQL.
func (bd *baseDialect) GetTenantSelectSQL() string {
	return fmt.Sprintf(selectTenantsSQL, migratorSchema, migratorTenantsTable)
}

// GetMigrationSelectSQL returns migrator's migrations select SQL statement.
// This SQL is used by all MySQL, PostgreSQL, MS SQL.
func (bd *baseDialect) GetMigrationSelectSQL() string {
	return fmt.Sprintf(selectMigrationsSQL, migratorSchema, migratorMigrationsTable)
}

// GetCreateSchemaSQL returns create schema SQL statement.
// This SQL is used by both MySQL and PostgreSQL.
func (bd *baseDialect) GetCreateSchemaSQL(schema string) string {
	return fmt.Sprintf(createSchemaSQL, schema)
}

// newDialect constructs dialect instance based on the passed Config
func newDialect(config *config.Config) (dialect, error) {

	var dialect dialect

	switch config.Driver {
	case "mysql":
		dialect = &mySQLDialect{}
	case "sqlserver":
		dialect = &msSQLDialect{}
	case "postgres":
		dialect = &postgreSQLDialect{}
	default:
		return nil, fmt.Errorf("Failed to create Connector: %q is an unknown driver", config.Driver)
	}

	return dialect, nil
}
