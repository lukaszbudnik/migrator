package db

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateDialectMSSQLDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "sqlserver"
	dialect := newDialect(config)
	assert.IsType(t, &msSQLDialect{}, dialect)
}

func TestMSSQLLastInsertIdSupported(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	dialect := newDialect(config)

	lastInsertIDSupported := dialect.LastInsertIDSupported()

	assert.False(t, lastInsertIDSupported)
}

func TestMSSQLGetMigrationInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := newDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_migrations (name, source_dir, filename, type, db_schema, contents, checksum, version_id) values (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)", insertMigrationSQL)
}

func TestMSSQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values (@p1)", tenantInsertSQL)
}

func TestMSSQLDialectGetCreateTenantsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := newDialect(config)

	createTenantsTableSQL := dialect.GetCreateTenantsTableSQL()

	expected := `
IF NOT EXISTS (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_tenants')
BEGIN
  create table [migrator].migrator_tenants (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP
  );
END
`

	assert.Equal(t, expected, createTenantsTableSQL)
}

func TestMSSQLDialectGetCreateMigrationsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := newDialect(config)

	createMigrationsTableSQL := dialect.GetCreateMigrationsTableSQL()

	expected := `
IF NOT EXISTS (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_migrations')
BEGIN
  create table [migrator].migrator_migrations (
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

	assert.Equal(t, expected, createMigrationsTableSQL)
}

func TestMSSQLDialectGetCreateSchemaSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := newDialect(config)

	createSchemaSQL := dialect.GetCreateSchemaSQL("def")

	expected := `
IF NOT EXISTS (select * from information_schema.schemata where schema_name = 'def')
BEGIN
  EXEC sp_executesql N'create schema def';
END
`

	assert.Equal(t, expected, createSchemaSQL)
}

func TestMSSQLGetVersionInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	dialect := newDialect(config)

	versionInsertSQL := dialect.GetVersionInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_versions (name) output inserted.id values (@p1)", versionInsertSQL)
}

func TestMSSQLGetCreateVersionsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	dialect := newDialect(config)

	actual := dialect.GetCreateVersionsTableSQL()

	expected :=
		`
if not exists (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_versions')
begin
  declare @cn nvarchar(200);
  create table [migrator].migrator_versions (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP
  );
  -- workaround for MSSQL not finding a newly created column
  -- when creating initial version default value is set to 1
  alter table [migrator].migrator_migrations add version_id int not null default 1;
  if exists (select * from [migrator].migrator_migrations)
  begin
    insert into [migrator].migrator_versions (name) values ('Initial version');
  end
  -- change version_id to not null
  alter table [migrator].migrator_migrations
    alter column version_id int not null;
  alter table [migrator].migrator_migrations
    add constraint migrator_versions_version_id_fk foreign key (version_id) references [migrator].migrator_versions (id) on delete cascade;
  create index migrator_migrations_version_id_idx on [migrator].migrator_migrations (version_id);
  -- remove workaround default value
  select @cn = name from sys.default_constraints where parent_object_id = object_id('[migrator].migrator_migrations') and name like '%ver%';
  EXEC ('alter table [migrator].migrator_migrations drop constraint ' + @cn);
end
`

	assert.Equal(t, expected, actual[0])
}

func TestMSSQLGetVersionsByFileSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	dialect := newDialect(config)

	versionsByFile := dialect.GetVersionsByFileSQL()

	assert.Equal(t, "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from migrator.migrator_versions mv left join migrator.migrator_migrations mm on mv.id = mm.version_id where mv.id in (select version_id from migrator.migrator_migrations where filename = @p1) order by vid desc, mid asc", versionsByFile)
}

func TestMSSQLGetVersionByIDSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	dialect := newDialect(config)

	versionByID := dialect.GetVersionByIDSQL()

	assert.Equal(t, "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from migrator.migrator_versions mv left join migrator.migrator_migrations mm on mv.id = mm.version_id where mv.id = @p1 order by mid asc", versionByID)
}

func TestMSSQLGetMigrationByIDSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	dialect := newDialect(config)

	migrationByID := dialect.GetMigrationByIDSQL()

	assert.Equal(t, "select id, name, source_dir, filename, type, db_schema, created, contents, checksum from migrator.migrator_migrations where id = @p1", migrationByID)
}
