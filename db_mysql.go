package main

import (
	_ "github.com/ziutek/mymysql/godrv"
)

type MySQLConnector struct {
	BaseConnector
}

const (
	insertMigrationMySQLDialectSQL = "insert into %v (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)"
)

func (mc *MySQLConnector) ApplyMigrations(migrations []Migration) error {
	return mc.BaseConnector.applyMigrationsWithInsertMigrationSQL(migrations, insertMigrationMySQLDialectSQL)
}
