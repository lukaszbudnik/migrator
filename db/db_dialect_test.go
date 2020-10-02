package db

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestBaseDialectGetCreateTenantsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)

	createTenantsTableSQL := dialect.GetCreateTenantsTableSQL()

	expected := `
create table if not exists migrator.migrator_tenants (
  id serial primary key,
  name varchar(200) not null,
  created timestamp default now()
)
`

	assert.Equal(t, expected, createTenantsTableSQL)
}

func TestBaseDialectGetCreateMigrationsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)

	createMigrationsTableSQL := dialect.GetCreateMigrationsTableSQL()

	expected := `
create table if not exists migrator.migrator_migrations (
  id serial primary key,
  name varchar(200) not null,
  source_dir varchar(200) not null,
  filename varchar(200) not null,
  type int not null,
  db_schema varchar(200) not null,
  created timestamp default now(),
	contents text,
	checksum varchar(64)
)
`

	assert.Equal(t, expected, createMigrationsTableSQL)
}

func TestBaseDialectGetCreateSchemaSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)

	createSchemaSQL := dialect.GetCreateSchemaSQL("abc")

	expected := "create schema if not exists abc"

	assert.Equal(t, expected, createSchemaSQL)
}

func TestBaseDialectGetVersionsSelectSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)

	versionsSelectSQL := dialect.GetVersionsSelectSQL()

	expected := "select mv.id as vid, mv.name as vname, mv.created as vcreated, mm.id as mid, mm.name, mm.source_dir, mm.filename, mm.type, mm.db_schema, mm.created, mm.contents, mm.checksum from migrator.migrator_versions mv left join migrator.migrator_migrations mm on mv.id = mm.version_id order by vid desc, mid asc"

	assert.Equal(t, expected, versionsSelectSQL)
}
