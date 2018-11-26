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
	insertMigrationPostgreSQLDialectSQL = "insert into %v (name, source_dir, file, type, db_schema) values ($1, $2, $3, $4, $5)"
	insertDefaultTenantPostgreSQLDialectSQL = "insert into %v (name) values ($1)"
)

func (pc *postgreSQLConnector) ApplyMigrations(migrations []types.Migration) {
	insertMigrationSQL := fmt.Sprintf(insertMigrationPostgreSQLDialectSQL, migrationsTableName)
	pc.BaseConnector.applyMigrationsWithInsertMigrationSQL(migrations, insertMigrationSQL)
}

func (pc *postgreSQLConnector) AddTenantAndApplyMigrations(tenant string, migrations []types.Migration) {
	insertMigrationSQL := fmt.Sprintf(insertMigrationPostgreSQLDialectSQL, migrationsTableName)
	insertTenantSQL := fmt.Sprintf(insertDefaultTenantPostgreSQLDialectSQL, defaultTenantsTableName)
	pc.BaseConnector.addTenantAndApplyMigrationsWithInsertTenantSQL(tenant, insertTenantSQL, migrations, insertMigrationSQL)
}
