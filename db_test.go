package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListAllDBTenants(t *testing.T) {
	config, err := readConfigFromFile("test/migrator.yaml")
	db, err := sql.Open(config.Driver, config.DataSource)

	assert.Nil(t, err)

	connector, err := CreateConnector(config.Driver)
	tenants, err := connector.ListAllDBTenants(*config, db)

	assert.Len(t, tenants, 3)
	assert.Equal(t, []string{"abc", "def", "xyz"}, tenants)

	db.Close()
}

func TestApplyMigrations(t *testing.T) {
	config, _ := readConfigFromFile("test/migrator.yaml")

	connector, err := CreateConnector(config.Driver)

	assert.Nil(t, err)

	allMigrations, _ := listAllMigrations(*config)
	dbMigrations, err := connector.ListAllDBMigrations(*config)

	assert.Nil(t, err)
	assert.Len(t, dbMigrations, 0)

	migrationDefs := computeMigrationsToApply(allMigrations, dbMigrations)
	migrations, _ := loadMigrations(*config, migrationDefs)

	err = connector.ApplyMigrations(*config, migrations)

	assert.Nil(t, err)

	dbMigrations, err = connector.ListAllDBMigrations(*config)

	assert.Nil(t, err)
	assert.Len(t, dbMigrations, 12)
}
