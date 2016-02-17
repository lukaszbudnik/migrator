package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

const (
	migrationsTableName     = "migrator_migrations"
	defaultTenantsTableName = "migrator_tenants"
	createMigrationsTable   = `
  create table if not exists %v (
    id serial primary key,
    name varchar(200) not null,
		source_dir varchar(200) not null,
    file varchar(200) not null,
		type int not null,
    schema varchar(200) not null,
    created timestamp default (now() at time zone 'utc')
	)
  `
	createDefaultTenantsTable = `
	create table if not exists %v (
		id serial primary key,
		name varchar(200) not null,
		created timestamp default (now() at time zone 'utc')
	)
	`
	selectMigrations     = "select distinct name, source_dir, file, type from %v order by name, source_dir"
	insertMigration      = "insert into %v (name, source_dir, file, type, schema) values ($1, $2, $3, $4, $5)"
	defaultSelectTenants = "select name from %v"
)

func listAllDBTenants(config Config, db *sql.DB) ([]string, error) {
	defaultTenantsSQL := fmt.Sprintf(defaultSelectTenants, defaultTenantsTableName)
	var tenantsSQL string
	if config.TenantsSQL != "" && config.TenantsSQL != defaultTenantsSQL {
		tenantsSQL = config.TenantsSQL
	} else {
		createTableQuery := fmt.Sprintf(createDefaultTenantsTable, defaultTenantsTableName)

		if _, err := db.Query(createTableQuery); err != nil {
			return nil, err
		}

		tenantsSQL = defaultTenantsSQL
	}

	rows, err := db.Query(tenantsSQL)
	if err != nil {
		return nil, err
	}
	var tenants []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tenants = append(tenants, name)
	}
	return tenants, nil
}

func listAllDBMigrations(config Config) ([]MigrationDefinition, error) {
	db, err := sql.Open(config.Driver, config.DataSource)
	if err != nil {
		return nil, err
	}

	createTableQuery := fmt.Sprintf(createMigrationsTable, migrationsTableName)
	if _, err := db.Query(createTableQuery); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(selectMigrations, migrationsTableName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	var dbMigrations []MigrationDefinition
	for rows.Next() {
		var (
			name          string
			sourceDir     string
			file          string
			migrationType MigrationType
		)
		if err := rows.Scan(&name, &sourceDir, &file, &migrationType); err != nil {
			return nil, err
		}
		dbMigrations = append(dbMigrations, MigrationDefinition{name, sourceDir, file, migrationType})
	}

	db.Close()

	return dbMigrations, err
}

func applyMigrations(config Config, migrations []Migration) error {

	if len(migrations) == 0 {
		return nil
	}

	db, err := sql.Open(config.Driver, config.DataSource)
	if err != nil {
		return err
	}

	tenants, err := listAllDBTenants(config, db)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := fmt.Sprintf(insertMigration, migrationsTableName)
	insert, err := db.Prepare(query)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		fmt.Println(m)
		var schemas []string
		if m.MigrationType == ModeTenantSchema {
			schemas = tenants
		} else {
			schemas = []string{m.SourceDir}
		}

		for _, s := range schemas {
			sql := strings.Replace(m.Contents, "{schema}", s, -1)
			_, err = tx.Exec(sql)
			if err != nil {
				tx.Rollback()
				return err
			}
			_, err = tx.Stmt(insert).Exec(m.Name, m.SourceDir, m.File, m.MigrationType, s)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

	}

	tx.Commit()

	return nil
}
