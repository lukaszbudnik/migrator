package db

import (
	"github.com/lukaszbudnik/migrator/types"
	// blank import for MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

type mySQLConnector struct {
	BaseConnector
}

const (
	insertMigrationMySQLDialectSQL = "insert into %v (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)"
)

func (mc *mySQLConnector) ApplyMigrations(migrations []types.Migration) {
	mc.BaseConnector.applyMigrationsWithInsertMigrationSQL(migrations, insertMigrationMySQLDialectSQL)
}
