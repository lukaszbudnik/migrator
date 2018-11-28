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
	insertMigrationMySQLDialectSql     = "insert into %v (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)"
	defaultInsertTenantMySQLDialectSql = "insert into %v (name) values (?)"
)

func (mc *mySQLConnector) ApplyMigrations(migrations []types.Migration) {
	insertMigrationSql := mc.GetMigrationInsertSql()
	mc.doApplyMigrations(migrations, insertMigrationSql)
}

func (mc *mySQLConnector) AddTenantAndApplyMigrations(tenant string, migrations []types.Migration) {
	insertMigrationSql := mc.GetMigrationInsertSql()
	insertTenantSql := mc.GetTenantInsertSql()
	mc.doAddTenantAndApplyMigrations(tenant, migrations, insertTenantSql, insertMigrationSql)
}

func (mc *mySQLConnector) GetMigrationInsertSql() string {
	return fmt.Sprintf(insertMigrationMySQLDialectSql, migrationsTableName)
}

func (mc *mySQLConnector) GetTenantInsertSql() string {
	var tenantsInsertSql string
	if mc.Config.TenantInsertSql != "" {
		tenantsInsertSql = mc.Config.TenantInsertSql
	} else {
		tenantsInsertSql = fmt.Sprintf(defaultInsertTenantMySQLDialectSql, defaultTenantsTableName)
	}
	return tenantsInsertSql
}
