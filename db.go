package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	migrationsTableName   = "migrator_migrations"
	createMigrationsTable = `
  create table if not exists %v (
    id serial primary key,
    name varchar(200) not null,
    schema varchar(200) not null,
    file varchar(200) not null,
    applied_at timestamp default (now() at time zone 'utc'))
  `
	selectMigrations = "select distinct file from %v order by file"
	insertMigration  = "insert into %v (name, schema, file) values ($1, $2, $3)"
)

func listAllDBMigrations(config Config) ([]string, error) {
	db, err := sql.Open(config.Driver, config.DataSource)

	createTableQuery := fmt.Sprintf(createMigrationsTable, migrationsTableName)

	if _, err := db.Query(createTableQuery); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(selectMigrations, migrationsTableName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var dbMigrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbMigrations = append(dbMigrations, name)
	}

	err = addMigrationEntry(db)

	db.Close()

	return dbMigrations, err
}

func addMigrationEntry(db *sql.DB) error {
	query := fmt.Sprintf(insertMigration, migrationsTableName)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	insert, err := db.Prepare(query)

	if err != nil {
		return err
	}

	_, err = tx.Stmt(insert).Exec("fofo", "gogo", "qqww.sql")

	tx.Commit()

	return err
}
