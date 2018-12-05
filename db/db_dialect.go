package db

import (
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"log"
)

type Dialect interface {
	GetTenantInsertSql() string
	GetTenantSelectSql() string
	GetMigrationInsertSql() string
	GetMigrationSelectSql() string
	GetCreateTenantsTableSql() string
	GetCreateMigrationsTableSql() string
	GetCreateSchemaSql(string) string
}

type BaseDialect struct {
}

const (
	selectMigrations = "select name, source_dir as sd, filename, type, db_schema, created from %v.%v order by name, source_dir"
	selectTenants    = "select name from %v.%v"
)

func (bd *BaseDialect) GetCreateTenantsTableSql() string {
	createTenantsTableSql := `
create table if not exists %v.%v (
  id serial primary key,
  name varchar(200) not null,
  created timestamp default now()
)
`
	return fmt.Sprintf(createTenantsTableSql, migratorSchema, migratorTenantsTable)
}

func (bd *BaseDialect) GetCreateMigrationsTableSql() string {
	createMigrationsTableSql := `
create table if not exists %v.%v (
  id serial primary key,
  name varchar(200) not null,
  source_dir varchar(200) not null,
  filename varchar(200) not null,
  type int not null,
  db_schema varchar(200) not null,
  created timestamp default now()
)
`
	return fmt.Sprintf(createMigrationsTableSql, migratorSchema, migratorMigrationsTable)
}

func (bd *BaseDialect) GetTenantSelectSql() string {
	return fmt.Sprintf(selectTenants, migratorSchema, migratorTenantsTable)
}

func (bd *BaseDialect) GetMigrationSelectSql() string {
	return fmt.Sprintf(selectMigrations, migratorSchema, migratorMigrationsTable)
}

func (bd *BaseDialect) GetCreateSchemaSql(schema string) string {
	return fmt.Sprintf("create schema if not exists %v", schema)
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
