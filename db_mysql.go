package main

import (
	_ "github.com/ziutek/mymysql/godrv"
)

type MySQLConnector struct {
	BaseConnector
}

const (
	insertMigrationMysql = "insert into %v (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)"
)

func (mc *MySQLConnector) ApplyMigrations(config Config, migrations []Migration) error {
	return mc.BaseConnector.applyMigrationsWithInsertMigration(config, migrations, insertMigrationMysql)
}
