package db

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateDialectMysqlDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "mysql"
	dialect := newDialect(config)
	assert.IsType(t, &mySQLDialect{}, dialect)
}

func TestMySQLLastInsertIdSupported(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)
	lastInsertIDSupported := dialect.LastInsertIDSupported()

	assert.True(t, lastInsertIDSupported)
}

func TestMySQLGetMigrationInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"

	dialect := newDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_migrations (name, source_dir, filename, type, db_schema, contents, checksum, version_id) values (?, ?, ?, ?, ?, ?, ?, ?)", insertMigrationSQL)
}

func TestMySQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values (?)", tenantInsertSQL)
}

func TestMySQLGetVersionInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)

	versionInsertSQL := dialect.GetVersionInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_versions (name) values (?)", versionInsertSQL)
}

func TestMySQLGetCreateVersionsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)

	actual := dialect.GetCreateVersionsTableSQL()
	expectedDrop := `drop procedure if exists migrator_create_versions`
	expectedCall := `call migrator_create_versions()`
	expectedProcedure :=
		`
create procedure migrator_create_versions()
begin
if not exists (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_versions') then
  create table migrator.migrator_versions (
    id serial primary key,
    name varchar(200) not null,
    created timestamp default now()
  );
  alter table migrator.migrator_migrations add column version_id bigint unsigned;
  create index migrator_versions_version_id_idx on migrator.migrator_migrations (version_id);
  if exists (select * from migrator.migrator_migrations) then
    insert into migrator.migrator_versions (name) values ('Initial version');
    -- initial version_id sequence is always 1
    update migrator.migrator_migrations set version_id = 1;
  end if;
  alter table migrator.migrator_migrations
    modify version_id bigint unsigned not null;
  alter table migrator.migrator_migrations
    add constraint migrator_versions_version_id_fk foreign key (version_id) references migrator.migrator_versions (id) on delete cascade;
end if;
end;
`

	assert.Equal(t, expectedDrop, actual[0])
	assert.Equal(t, expectedProcedure, actual[1])
	assert.Equal(t, expectedCall, actual[2])
}

func TestMySQLGetVersionsByFileSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)

	versionsByFile := dialect.GetVersionsByFileSQL()

	assert.Equal(t, "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from migrator.migrator_versions mv left join migrator.migrator_migrations mm on mv.id = mm.version_id where mv.id in (select version_id from migrator.migrator_migrations where filename = ?) order by vid desc, mid asc", versionsByFile)
}

func TestMySQLGetVersionByIDSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)

	versionsByID := dialect.GetVersionByIDSQL()

	assert.Equal(t, "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from migrator.migrator_versions mv left join migrator.migrator_migrations mm on mv.id = mm.version_id where mv.id = ? order by mid asc", versionsByID)
}

func TestMySQLGetMigrationByIDSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)

	migrationByID := dialect.GetMigrationByIDSQL()

	assert.Equal(t, "select id, name, source_dir, filename, type, db_schema, created, contents, checksum from migrator.migrator_migrations where id = ?", migrationByID)
}
