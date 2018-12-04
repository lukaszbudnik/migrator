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
	insertMigrationMySQLDialectSql = "insert into %v.%v (name, source_dir, filename, type, db_schema) values (?, ?, ?, ?, ?)"
	insertTenantMySQLDialectSql    = "insert into %v.%v (name) values (?)"
)

func (md *mySQLDialect) GetMigrationInsertSql() string {
	return fmt.Sprintf(insertMigrationMySQLDialectSql, migratorSchema, migratorMigrationsTable)
}

func (md *mySQLDialect) GetTenantInsertSql() string {
	return fmt.Sprintf(insertTenantMySQLDialectSql, migratorSchema, migratorTenantsTable)
}
