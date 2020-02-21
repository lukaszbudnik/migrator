package db

import (
	"fmt"
	// blank import for MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

type mySQLDialect struct {
	baseDialect
}

const (
	insertMigrationMySQLDialectSQL             = "insert into %v.%v (name, source_dir, filename, type, db_schema, contents, checksum, version_id) values (?, ?, ?, ?, ?, ?, ?, ?)"
	insertTenantMySQLDialectSQL                = "insert into %v.%v (name) values (?)"
	insertVersionMySQLDialectSQL               = "insert into %v.%v (name) values (?)"
	versionsTableSetupMySQLDropDialectSQL      = `drop procedure if exists migrator_create_versions`
	versionsTableSetupMySQLCallDialectSQL      = `call migrator_create_versions()`
	versionsTableSetupMySQLProcedureDialectSQL = `
create procedure migrator_create_versions()
begin
if not exists (select * from information_schema.tables where table_schema = '%v' and table_name = '%v') then
  create table %v.%v (
    id serial primary key,
    name varchar(200) not null,
    created timestamp default now()
  );
  alter table %v.%v add column version_id bigint unsigned;
  create index migrator_versions_version_id_idx on %v.%v (version_id);
  if exists (select * from %v.%v) then
    insert into %v.%v (name) values ('Initial version');
    -- initial version_id sequence is always 1
    update %v.%v set version_id = 1;
  end if;
  alter table %v.%v
    modify version_id bigint unsigned not null;
  alter table %v.%v
    add constraint migrator_versions_version_id_fk foreign key (version_id) references %v.%v (id) on delete cascade;
end if;
end;
`
)

// GetMigrationInsertSQL returns MySQL-specific migration insert SQL statement
func (md *mySQLDialect) GetMigrationInsertSQL() string {
	return fmt.Sprintf(insertMigrationMySQLDialectSQL, migratorSchema, migratorMigrationsTable)
}

// GetTenantInsertSQL returns MySQL-specific migrator's default tenant insert SQL statement
func (md *mySQLDialect) GetTenantInsertSQL() string {
	return fmt.Sprintf(insertTenantMySQLDialectSQL, migratorSchema, migratorTenantsTable)
}

func (md *mySQLDialect) GetVersionInsertSQL() string {
	return fmt.Sprintf(insertVersionMySQLDialectSQL, migratorSchema, migratorVersionsTable)
}

// GetCreateVersionsTableSQL returns MySQL-specific SQLs which does:
// 1. drop procedure if exists
// 2. create procedure
// 3. calls procedure
// far from ideal MySQL in contrast to MS SQL and PostgreSQL does not support the execution of anonymous blocks of code
func (md *mySQLDialect) GetCreateVersionsTableSQL() []string {
	return []string{
		versionsTableSetupMySQLDropDialectSQL,
		fmt.Sprintf(versionsTableSetupMySQLProcedureDialectSQL, migratorSchema, migratorVersionsTable, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorVersionsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorMigrationsTable, migratorSchema, migratorVersionsTable),
		versionsTableSetupMySQLCallDialectSQL,
	}
}
