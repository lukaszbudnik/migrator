package db

import (
	"github.com/lukaszbudnik/migrator/types"
	_ "github.com/ziutek/mymysql/godrv"
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
