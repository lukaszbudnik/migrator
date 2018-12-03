package db

import (
	"github.com/lukaszbudnik/migrator/config"
	"log"
)

type Dialect interface {
	GetTenantInsertSql() string
	GetMigrationInsertSql() string
  GetCreateTenantsTableSql() string
  GetCreateMigrationsTableSql() string
}

type BaseDialect struct {
}

func (bd *BaseDialect) GetCreateTenantsTableSql() string {
  return `
  create table if not exists %v (
    id serial primary key,
    name varchar(200) not null,
    created timestamp default now()
  );
  `
}

func (bd *BaseDialect) GetCreateMigrationsTableSql() string {
  return `
  create table if not exists %v (
    id serial primary key,
    name varchar(200) not null,
    source_dir varchar(200) not null,
    file varchar(200) not null,
    type int not null,
    db_schema varchar(200) not null,
    created timestamp default now()
  );
  `
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
