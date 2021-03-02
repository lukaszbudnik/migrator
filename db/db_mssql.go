package db

import (
	"fmt"
	// blank import for MSSQL driver
	_ "github.com/denisenkom/go-mssqldb"
)

type msSQLDialect struct {
	baseDialect
}

const (
	insertMigrationMSSQLDialectSQL      = "insert into %v.%v (name, source_dir, filename, type, db_schema, contents, checksum, version_id) values (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)"
	insertTenantMSSQLDialectSQL         = "insert into %v.%v (name) values (@p1)"
	insertVersionMSSQLSQLDialectSQL     = "insert into %v.%v (name) output inserted.id values (@p1)"
	selectVersionsByFileMSSQLDialectSQL = "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from %v.%v mv left join %v.%v mm on mv.id = mm.version_id where mv.id in (select version_id from %v.%v where filename = @p1) order by vid desc, mid asc"
	selectVersionByIDMSSQLDialectSQL    = "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from %v.%v mv left join %v.%v mm on mv.id = mm.version_id where mv.id = @p1 order by mid asc"
	selectMigrationByIDMSSQLDialectSQL  = "select id, name, source_dir, filename, type, db_schema, created, contents, checksum from %v.%v where id = @p1"
	createTenantsTableMSSQLDialectSQL   = `
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
	versionsTableSetupMSSQLDialectSQL = `
if not exists (select * from information_schema.tables where table_schema = '%v' and table_name = '%v')
begin
  declare @cn nvarchar(200);
  create table [%v].%v (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP
  );
  -- workaround for MSSQL not finding a newly created column
  -- when creating initial version default value is set to 1
  alter table [%v].%v add version_id int not null default 1;
  if exists (select * from [%v].%v)
  begin
    insert into [%v].%v (name) values ('Initial version');
  end
  -- change version_id to not null
  alter table [%v].%v
    alter column version_id int not null;
  alter table [%v].%v
    add constraint migrator_versions_version_id_fk foreign key (version_id) references [%v].%v (id) on delete cascade;
  create index migrator_migrations_version_id_idx on [%v].%v (version_id);
  -- remove workaround default value
  select @cn = name from sys.default_constraints where parent_object_id = object_id('[%v].%v') and name like '%%ver%%';
  EXEC ('alter table [%v].%v drop constraint ' + @cn);
end
`
)

// LastInsertIDSupported instructs migrator if Result.LastInsertId() is supported by the DB driver
func (md *msSQLDialect) LastInsertIDSupported() bool {
	return false
}

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
// This SQL is used by MS SQL.
func (md *msSQLDialect) GetCreateMigrationsTableSQL() string {
	return fmt.Sprintf(createMigrationsTableMSSQLDialectSQL, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable)
}

// GetCreateSchemaSQL returns create schema SQL statement.
// This SQL is used by MS SQL.
func (md *msSQLDialect) GetCreateSchemaSQL(schema string) string {
	if !isValidIdentifier(schema) {
		panic(fmt.Sprintf("Schema name contains invalid characters: %v", schema))
	}
	return fmt.Sprintf(createSchemaMSSQLDialectSQL, schema, schema)
}

func (md *msSQLDialect) GetVersionInsertSQL() string {
	return fmt.Sprintf(insertVersionMSSQLSQLDialectSQL, migratorSchema, migratorVersionsTable)
}

func (md *msSQLDialect) GetCreateVersionsTableSQL() []string {
	return []string{fmt.Sprintf(versionsTableSetupMSSQLDialectSQL, migratorSchema, migratorVersionsTable, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable)}
}

func (md *msSQLDialect) GetVersionsByFileSQL() string {
	return fmt.Sprintf(selectVersionsByFileMSSQLDialectSQL, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable)
}

func (md *msSQLDialect) GetVersionByIDSQL() string {
	return fmt.Sprintf(selectVersionByIDMSSQLDialectSQL, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable)
}

func (md *msSQLDialect) GetMigrationByIDSQL() string {
	return fmt.Sprintf(selectMigrationByIDMSSQLDialectSQL, migratorSchema, migratorMigrationsTable)
}
