package db

import (
	"fmt"
	// blank import for MSSQL driver
	_ "github.com/denisenkom/go-mssqldb"
)

type msSQLDialect struct {
	BaseDialect
}

const (
	insertMigrationMSSQLDialectSQL    = "insert into %v.%v (name, source_dir, filename, type, db_schema, contents, checksum) values (@p1, @p2, @p3, @p4, @p5, @p6, @p7)"
	insertTenantMSSQLDialectSQL       = "insert into %v.%v (name) values (@p1)"
	createTenantsTableMSSQLDialectSQL = `
IF NOT EXISTS (select * from information_schema.tables where table_schema = '%v' and table_name = '%v')
BEGIN
  create table [%v].%v (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP
  );
END
`
	createMigrationsTableMSSQLDialectSQL = `
IF NOT EXISTS (select * from information_schema.tables where table_schema = '%v' and table_name = '%v')
BEGIN
  create table [%v].%v (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    source_dir varchar(200) not null,
    filename varchar(200) not null,
    type int not null,
    db_schema varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP,
		contents text,
		checksum varchar(64)
  );
END
`
	createSchemaMSSQLDialectSQL = `
IF NOT EXISTS (select * from information_schema.schemata where schema_name = '%v')
BEGIN
  EXEC sp_executesql N'create schema %v';
END
`
)

// GetMigrationInsertSQL returns MS SQL-specific migration insert SQL statement
func (md *msSQLDialect) GetMigrationInsertSQL() string {
	return fmt.Sprintf(insertMigrationMSSQLDialectSQL, migratorSchema, migratorMigrationsTable)
}

// GetTenantInsertSQL returns MS SQL-specific migrator's default tenant insert SQL statement
func (md *msSQLDialect) GetTenantInsertSQL() string {
	return fmt.Sprintf(insertTenantMSSQLDialectSQL, migratorSchema, migratorTenantsTable)
}

// GetCreateTenantsTableSQL returns migrator's default create tenants table SQL statement.
// This SQL is used by MS SQL.
func (md *msSQLDialect) GetCreateTenantsTableSQL() string {
	return fmt.Sprintf(createTenantsTableMSSQLDialectSQL, migratorSchema, migratorTenantsTable, migratorSchema, migratorTenantsTable)
}

// GetCreateMigrationsTableSQL returns migrator's create migrations table SQL statement.
// This SQL is used by both MS SQL.
func (md *msSQLDialect) GetCreateMigrationsTableSQL() string {
	return fmt.Sprintf(createMigrationsTableMSSQLDialectSQL, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable)
}

// GetCreateSchemaSQL returns create schema SQL statement.
// This SQL is used by both MySQL and PostgreSQL.
func (md *msSQLDialect) GetCreateSchemaSQL(schema string) string {
	return fmt.Sprintf(createSchemaMSSQLDialectSQL, schema, schema)
}
