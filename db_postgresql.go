package main

import (
	_ "github.com/lib/pq"
)

type PostgresqlConnector struct {
	BaseConnector
}

const (
	insertMigrationPostgresql = "insert into %v (name, source_dir, file, type, db_schema) values ($1, $2, $3, $4, $5)"
)

func (pc *PostgresqlConnector) ApplyMigrations(config Config, migrations []Migration) error {
	return pc.BaseConnector.applyMigrationsWithInsertMigration(config, migrations, insertMigrationPostgresql)
}
