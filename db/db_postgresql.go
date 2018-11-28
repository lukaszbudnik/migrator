package db

import (
	"fmt"
	"github.com/lukaszbudnik/migrator/types"
	// blank import for PostgreSQL driver
	_ "github.com/lib/pq"
)

type postgreSQLConnector struct {
	BaseConnector
}

const (
	insertMigrationPostgreSQLDialectSql     = "insert into %v (name, source_dir, file, type, db_schema) values ($1, $2, $3, $4, $5)"
	defaultInsertTenantPostgreSQLDialectSql = "insert into %v (name) values ($1)"
)

func (pc *postgreSQLConnector) ApplyMigrations(migrations []types.Migration) {
	insertMigrationSql := pc.GetMigrationInsertSql()
	pc.doApplyMigrations(migrations, insertMigrationSql)
}

func (pc *postgreSQLConnector) AddTenantAndApplyMigrations(tenant string, migrations []types.Migration) {
	insertMigrationSql := pc.GetMigrationInsertSql()
	insertTenantSql := pc.GetTenantInsertSql()
	pc.doAddTenantAndApplyMigrations(tenant, migrations, insertTenantSql, insertMigrationSql)
}

func (pc *postgreSQLConnector) GetMigrationInsertSql() string {
	return fmt.Sprintf(insertMigrationPostgreSQLDialectSql, migrationsTableName)
}

func (pc *postgreSQLConnector) GetTenantInsertSql() string {
	var tenantsInsertSql string
	if pc.Config.TenantInsertSql != "" {
		tenantsInsertSql = pc.Config.TenantInsertSql
	} else {
		tenantsInsertSql = fmt.Sprintf(defaultInsertTenantPostgreSQLDialectSql, defaultTenantsTableName)
	}
	return tenantsInsertSql
}
