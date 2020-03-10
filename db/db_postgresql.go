package db

import (
	"fmt"
	// blank import for PostgreSQL driver
	_ "github.com/lib/pq"
)

type postgreSQLDialect struct {
	baseDialect
}

const (
	insertMigrationPostgreSQLDialectSQL      = "insert into %v.%v (name, source_dir, filename, type, db_schema, contents, checksum, version_id) values ($1, $2, $3, $4, $5, $6, $7, $8)"
	insertTenantPostgreSQLDialectSQL         = "insert into %v.%v (name) values ($1)"
	insertVersionPostgreSQLDialectSQL        = "insert into %v.%v (name) values ($1) returning id"
	selectVersionsByFilePostgreSQLDialectSQL = "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from %v.%v mv left join %v.%v mm on mv.id = mm.version_id where mv.id in (select version_id from %v.%v where filename = $1) order by vid desc, mid asc"
	selectVersionByIDPostgreSQLDialectSQL    = "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from %v.%v mv left join %v.%v mm on mv.id = mm.version_id where mv.id = $1 order by mid asc"
	selectMigrationByIDPostgreSQLDialectSQL  = "select id, name, source_dir, filename, type, db_schema, created, contents, checksum from %v.%v where id = $1"
	versionsTableSetupPostgreSQLDialectSQL   = `
do $$
begin
if not exists (select * from information_schema.tables where table_schema = '%v' and table_name = '%v') then
  create table %v.%v (
    id serial primary key,
    name varchar(200) not null,
    created timestamp default now()
  );
  alter table %v.%v add column version_id integer;
  create index migrator_versions_version_id_idx on %v.%v (version_id);
  if exists (select * from %v.%v) then
    insert into %v.%v (name) values ('Initial version');
    -- initial version_id sequence is always 1
    update %v.%v set version_id = 1;
  end if;
  alter table %v.%v
    alter column version_id set not null,
    add constraint migrator_versions_version_id_fk foreign key (version_id) references %v.%v (id) on delete cascade;
end if;
end $$;
`
)

// LastInsertIDSupported instructs migrator if Result.LastInsertId() is supported by the DB driver
func (pd *postgreSQLDialect) LastInsertIDSupported() bool {
	return false
}

// GetMigrationInsertSQL returns PostgreSQL-specific migration insert SQL statement
func (pd *postgreSQLDialect) GetMigrationInsertSQL() string {
	return fmt.Sprintf(insertMigrationPostgreSQLDialectSQL, migratorSchema, migratorMigrationsTable)
}

// GetTenantInsertSQL returns PostgreSQL-specific migrator's default tenant insert SQL statement
func (pd *postgreSQLDialect) GetTenantInsertSQL() string {
	return fmt.Sprintf(insertTenantPostgreSQLDialectSQL, migratorSchema, migratorTenantsTable)
}

func (pd *postgreSQLDialect) GetVersionInsertSQL() string {
	return fmt.Sprintf(insertVersionPostgreSQLDialectSQL, migratorSchema, migratorVersionsTable)
}

// GetCreateVersionsTableSQL returns PostgreSQL-specific SQL which does:
// 1. create versions table
// 2. alter statement used to add version column to migration
// 3. create initial version if migrations exists (backwards compatibility)
// 4. create not null consttraint on version column
func (pd *postgreSQLDialect) GetCreateVersionsTableSQL() []string {
	return []string{fmt.Sprintf(versionsTableSetupPostgreSQLDialectSQL, migratorSchema, migratorVersionsTable, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorVersionsTable)}
}

func (pd *postgreSQLDialect) GetVersionsByFileSQL() string {
	return fmt.Sprintf(selectVersionsByFilePostgreSQLDialectSQL, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable)
}

func (pd *postgreSQLDialect) GetVersionByIDSQL() string {
	return fmt.Sprintf(selectVersionByIDPostgreSQLDialectSQL, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable)
}

func (pd *postgreSQLDialect) GetMigrationByIDSQL() string {
	return fmt.Sprintf(selectMigrationByIDPostgreSQLDialectSQL, migratorSchema, migratorMigrationsTable)
}
