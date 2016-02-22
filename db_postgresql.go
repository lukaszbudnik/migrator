package main

import (
	_ "github.com/lib/pq"
)

type postgresqlConnector struct {
	BaseConnector
}

const (
	insertMigrationPostgreSQLDialectSQL = "insert into %v (name, source_dir, file, type, db_schema) values ($1, $2, $3, $4, $5)"
)

func (pc *postgresqlConnector) ApplyMigrations(migrations []Migration) error {
	return pc.BaseConnector.applyMigrationsWithInsertMigrationSQL(migrations, insertMigrationPostgreSQLDialectSQL)
}
