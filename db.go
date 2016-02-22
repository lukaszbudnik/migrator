package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// Connector interface abstracts all DB operations performed by migrator
type Connector interface {
	Init()
	ListAllDBTenants() ([]string, error)
	ListAllDBMigrations() ([]DBMigration, error)
	ApplyMigrations(migrations []Migration) error
	Dispose()
}

// BaseConnector struct is a base struct for implementing DB specific dialects
type BaseConnector struct {
	Config *Config
	DB     *sql.DB
}

// CreateConnector constructs Connector instance based on the passed Config
func CreateConnector(config *Config) Connector {
	var bc = BaseConnector{config, nil}
	var connector Connector

	switch config.Driver {
	case "mymysql":
		connector = &mySQLConnector{bc}
	case "postgres":
		connector = &postgresqlConnector{bc}
	default:
		log.Panicf("Failed to create Connector: %q is an unknown driver.", config.Driver)
	}

	connector.Init()

	return connector
}

const (
	migrationsTableName     = "public.migrator_migrations"
	defaultTenantsTableName = "public.migrator_tenants"
	createMigrationsTable   = `
  create table if not exists %v (
    id serial primary key,
    name varchar(200) not null,
		source_dir varchar(200) not null,
    file varchar(200) not null,
		type int not null,
    db_schema varchar(200) not null,
    created timestamp default now()
	)
  `
	createDefaultTenantsTable = `
	create table if not exists %v (
		id serial primary key,
		name varchar(200) not null,
		created timestamp default now()
	)
	`
	selectMigrations     = "select distinct name, source_dir, file, type, db_schema, created from %v order by name, source_dir"
	defaultSelectTenants = "select name from %v"
)

// Init initialises connector by opening a connection to database
func (bc *BaseConnector) Init() {
	db, err := sql.Open(bc.Config.Driver, bc.Config.DataSource)
	if err != nil {
		log.Panicf("Failed to init base connector ==> %v", err)
	}
	bc.DB = db
}

// Dispose closes all resources allocated by connector
func (bc *BaseConnector) Dispose() {
	bc.DB.Close()
}

// ListAllDBTenants returns a list of all DB tenants as specified by
// defaultSelectTenants or the value specified in config
func (bc *BaseConnector) ListAllDBTenants() ([]string, error) {
	defaultTenantsSQL := fmt.Sprintf(defaultSelectTenants, defaultTenantsTableName)
	var tenantsSQL string
	if bc.Config.TenantsSQL != "" && bc.Config.TenantsSQL != defaultTenantsSQL {
		tenantsSQL = bc.Config.TenantsSQL
	} else {
		createTableQuery := fmt.Sprintf(createDefaultTenantsTable, defaultTenantsTableName)

		if _, err := bc.DB.Query(createTableQuery); err != nil {
			return nil, err
		}

		tenantsSQL = defaultTenantsSQL
	}

	rows, err := bc.DB.Query(tenantsSQL)
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

// ListAllDBMigrations returns a list of all applied DB migrations
func (bc *BaseConnector) ListAllDBMigrations() ([]DBMigration, error) {
	createTableQuery := fmt.Sprintf(createMigrationsTable, migrationsTableName)
	if _, err := bc.DB.Query(createTableQuery); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(selectMigrations, migrationsTableName)

	rows, err := bc.DB.Query(query)
	if err != nil {
		return nil, err
	}

	var dbMigrations []DBMigration
	for rows.Next() {
		var (
			name          string
			sourceDir     string
			file          string
			migrationType MigrationType
			schema        string
			created       time.Time
		)
		if err := rows.Scan(&name, &sourceDir, &file, &migrationType, &schema, &created); err != nil {
			return nil, err
		}
		mdef := MigrationDefinition{name, sourceDir, file, migrationType}
		dbMigrations = append(dbMigrations, DBMigration{mdef, schema, created})
	}

	log.Println("Read db migrations ==> OK")

	return dbMigrations, nil
}

// ApplyMigrations applies passed migrations
func (bc *BaseConnector) ApplyMigrations(migrations []Migration) error {
	panic("ApplyMigrations() must be overwritten by specific connector")
}

// applyMigrationsWithInsertMigrationSQL is called by specific implementations
// insertMigrationSQL varies based on database dialect
func (bc *BaseConnector) applyMigrationsWithInsertMigrationSQL(migrations []Migration, insertMigrationSQL string) error {

	if len(migrations) == 0 {
		return nil
	}

	tenants, err := bc.ListAllDBTenants()
	if err != nil {
		return err
	}

	tx, err := bc.DB.Begin()
	if err != nil {
		return err
	}

	query := fmt.Sprintf(insertMigrationSQL, migrationsTableName)
	insert, err := bc.DB.Prepare(query)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		var schemas []string
		if m.MigrationType == ModeTenantSchema {
			schemas = tenants
		} else {
			schemas = []string{m.SourceDir}
		}

		for _, s := range schemas {
			contents := strings.Replace(m.Contents, "{schema}", s, -1)
			sqls := strings.Split(contents, ";")
			for _, sql := range sqls {
				if strings.TrimSpace(sql) != "" {
					_, err = tx.Exec(sql)
					if err != nil {
						tx.Rollback()
						return err
					}
				}
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
