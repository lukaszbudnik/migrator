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
	insertMigrationMySQLDialectSql     = "insert into %v (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)"
	defaultInsertTenantMySQLDialectSql = "insert into %v (name) values (?)"
)

func (md *mySQLDialect) GetMigrationInsertSql() string {
	return fmt.Sprintf(insertMigrationMySQLDialectSql, migrationsTableName)
}

func (md *mySQLDialect) GetTenantInsertSql() string {
	return fmt.Sprintf(defaultInsertTenantMySQLDialectSql, defaultTenantsTableName)
}
