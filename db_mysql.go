package main

import (
	_ "github.com/ziutek/mymysql/godrv"
)

type mySQLConnector struct {
	BaseConnector
}

const (
	insertMigrationMySQLDialectSQL = "insert into %v (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)"
)

func (mc *mySQLConnector) ApplyMigrations(migrations []Migration) error {
	return mc.BaseConnector.applyMigrationsWithInsertMigrationSQL(migrations, insertMigrationMySQLDialectSQL)
}
