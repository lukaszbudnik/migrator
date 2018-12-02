package db

import (
	"github.com/lukaszbudnik/migrator/config"
	"log"
)

type Dialect interface {
	GetTenantInsertSql() string
	GetMigrationInsertSql() string
}

type BaseDialect struct {
}

// CreateDialect constructs Dialect instance based on the passed Config
func CreateDialect(config *config.Config) Dialect {

	var dialect Dialect

	switch config.Driver {
	case "mysql":
		dialect = &mySQLDialect{}
	case "postgres":
		dialect = &postgreSQLDialect{}
	default:
		log.Panicf("Failed to create Connector: %q is an unknown driver.", config.Driver)
	}

	return dialect
}
