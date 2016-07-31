package db

import (
	// blank import for PostgreSQL driver
	_ "github.com/lib/pq"
	"github.com/lukaszbudnik/migrator/types"
)

type postgreSQLConnector struct {
	BaseConnector
}

const (
	insertMigrationPostgreSQLDialectSQL = "insert into %v (name, source_dir, file, type, db_schema) values ($1, $2, $3, $4, $5)"
)

func (pc *postgreSQLConnector) ApplyMigrations(migrations []types.Migration) {
	pc.BaseConnector.applyMigrationsWithInsertMigrationSQL(migrations, insertMigrationPostgreSQLDialectSQL)
}
