package db

import (
	"fmt"
	// blank import for MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

type mySQLDialect struct {
	BaseDialect
}

const (
	insertMigrationMySQLDialectSQL = "insert into %v.%v (name, source_dir, filename, type, db_schema, contents, checksum) values (?, ?, ?, ?, ?, ?, ?)"
	insertTenantMySQLDialectSQL    = "insert into %v.%v (name) values (?)"
)

// GetMigrationInsertSQL returns MySQL-specific migration insert SQL statement
func (md *mySQLDialect) GetMigrationInsertSQL() string {
	return fmt.Sprintf(insertMigrationMySQLDialectSQL, migratorSchema, migratorMigrationsTable)
}

// GetTenantInsertSQL returns MySQL-specific migrator's default tenant insert SQL statement
func (md *mySQLDialect) GetTenantInsertSQL() string {
	return fmt.Sprintf(insertTenantMySQLDialectSQL, migratorSchema, migratorTenantsTable)
}
