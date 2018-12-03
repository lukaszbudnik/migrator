package db

import (
	"fmt"
	// blank import for PostgreSQL driver
	_ "github.com/lib/pq"
)

type postgreSQLDialect struct {
	BaseDialect
}

const (
	insertMigrationPostgreSQLDialectSql     = "insert into %v (name, source_dir, file, type, db_schema) values ($1, $2, $3, $4, $5)"
	defaultInsertTenantPostgreSQLDialectSql = "insert into %v (name) values ($1)"
)

func (pd *postgreSQLDialect) GetMigrationInsertSql() string {
	return fmt.Sprintf(insertMigrationPostgreSQLDialectSql, migrationsTableName)
}

func (pd *postgreSQLDialect) GetTenantInsertSql() string {
	return fmt.Sprintf(defaultInsertTenantPostgreSQLDialectSql, defaultTenantsTableName)
}
