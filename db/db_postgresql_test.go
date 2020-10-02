package db

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateDialectPostgreSQLDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	assert.IsType(t, &postgreSQLDialect{}, dialect)
}

func TestPostgreSQLLastInsertIdSupported(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	dialect := newDialect(config)
	lastInsertIDSupported := dialect.LastInsertIDSupported()

	assert.False(t, lastInsertIDSupported)
}

func TestPostgreSQLGetMigrationInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := newDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_migrations (name, source_dir, filename, type, db_schema, contents, checksum, version_id) values ($1, $2, $3, $4, $5, $6, $7, $8)", insertMigrationSQL)
}

func TestPostgreSQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values ($1)", tenantInsertSQL)
}

func TestPostgreSQLGetVersionInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	dialect := newDialect(config)

	versionInsertSQL := dialect.GetVersionInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_versions (name) values ($1) returning id", versionInsertSQL)
}

func TestPostgreSQLGetCreateVersionsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	dialect := newDialect(config)

	actual := dialect.GetCreateVersionsTableSQL()

	expected :=
		`
do $$
begin
if not exists (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_versions') then
  create table migrator.migrator_versions (
    id serial primary key,
    name varchar(200) not null,
    created timestamp with time zone default now()
  );
  alter table migrator.migrator_migrations add column version_id integer;
  create index migrator_versions_version_id_idx on migrator.migrator_migrations (version_id);
  if exists (select * from migrator.migrator_migrations) then
    insert into migrator.migrator_versions (name) values ('Initial version');
    -- initial version_id sequence is always 1
    update migrator.migrator_migrations set version_id = 1;
  end if;
  alter table migrator.migrator_migrations
    alter column version_id set not null,
    add constraint migrator_versions_version_id_fk foreign key (version_id) references migrator.migrator_versions (id) on delete cascade;
end if;
end $$;
`

	assert.Equal(t, expected, actual[0])
}

func TestPostgreSQLGetVersionsByFileSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	dialect := newDialect(config)

	versionsByFile := dialect.GetVersionsByFileSQL()

	assert.Equal(t, "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from migrator.migrator_versions mv left join migrator.migrator_migrations mm on mv.id = mm.version_id where mv.id in (select version_id from migrator.migrator_migrations where filename = $1) order by vid desc, mid asc", versionsByFile)
}

func TestPostgreSQLGetVersionByIDSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	dialect := newDialect(config)

	versionsByID := dialect.GetVersionByIDSQL()

	assert.Equal(t, "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from migrator.migrator_versions mv left join migrator.migrator_migrations mm on mv.id = mm.version_id where mv.id = $1 order by mid asc", versionsByID)
}

func TestPostgreSQLGetMigrationByIDSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	dialect := newDialect(config)

	migrationByID := dialect.GetMigrationByIDSQL()

	assert.Equal(t, "select id, name, source_dir, filename, type, db_schema, created, contents, checksum from migrator.migrator_migrations where id = $1", migrationByID)
}
