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
	insertMigrationPostgreSQLDialectSQL = "insert into %v.%v (name, source_dir, filename, type, db_schema) values ($1, $2, $3, $4, $5)"
	insertTenantPostgreSQLDialectSQL    = "insert into %v.%v (name) values ($1)"
)

// GetMigrationInsertSQL returns PostgreSQL-specific migration insert SQL statement
func (pd *postgreSQLDialect) GetMigrationInsertSQL() string {
	return fmt.Sprintf(insertMigrationPostgreSQLDialectSQL, migratorSchema, migratorMigrationsTable)
}

// GetTenantInsertSQL returns PostgreSQL-specific migrator's default tenant insert SQL statement
func (pd *postgreSQLDialect) GetTenantInsertSQL() string {
	return fmt.Sprintf(insertTenantPostgreSQLDialectSQL, migratorSchema, migratorTenantsTable)
}
