package db

import (
	"fmt"
	"github.com/lukaszbudnik/migrator/types"
	// blank import for MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

type mySQLConnector struct {
	BaseConnector
}

const (
	insertMigrationMySQLDialectSQL = "insert into %v (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)"
	insertDefaultTenantMySQLDialectSQL = "insert into %v (name) values (?)"
)

func (mc *mySQLConnector) ApplyMigrations(migrations []types.Migration) {
	insertMigrationSQL := fmt.Sprintf(insertMigrationMySQLDialectSQL, migrationsTableName)
	mc.BaseConnector.applyMigrationsWithInsertMigrationSQL(migrations, insertMigrationSQL)
}

func (mc *mySQLConnector) AddTenantAndApplyMigrations(tenant string, migrations []types.Migration) {
	insertMigrationSQL := fmt.Sprintf(insertMigrationMySQLDialectSQL, migrationsTableName)
	insertTenantSQL := fmt.Sprintf(insertDefaultTenantMySQLDialectSQL, defaultTenantsTableName)
	mc.BaseConnector.addTenantAndApplyMigrationsWithInsertTenantSQL(tenant, insertTenantSQL, migrations, insertMigrationSQL)
}
