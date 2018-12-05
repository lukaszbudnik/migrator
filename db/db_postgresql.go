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
	insertMigrationPostgreSQLDialectSql = "insert into %v.%v (name, source_dir, filename, type, db_schema) values ($1, $2, $3, $4, $5)"
	insertTenantPostgreSQLDialectSql    = "insert into %v.%v (name) values ($1)"
)

func (pd *postgreSQLDialect) GetMigrationInsertSql() string {
	return fmt.Sprintf(insertMigrationPostgreSQLDialectSql, migratorSchema, migratorMigrationsTable)
}

func (pd *postgreSQLDialect) GetTenantInsertSql() string {
	return fmt.Sprintf(insertTenantPostgreSQLDialectSql, migratorSchema, migratorTenantsTable)
}
