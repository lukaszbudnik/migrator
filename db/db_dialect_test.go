package db

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestBaseDialectGetCreateTenantsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

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
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

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
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := newDialect(config)

	createSchemaSQL := dialect.GetCreateSchemaSQL("abc")

	expected := "create schema if not exists abc"

	assert.Equal(t, expected, createSchemaSQL)
}

func TestBaseDialectGetVersionsSelectSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := newDialect(config)

	versionsSelectSQL := dialect.GetVersionsSelectSQL()

	expected := "select id, name, created from migrator.migrator_versions order by created desc"

	assert.Equal(t, expected, versionsSelectSQL)
}
