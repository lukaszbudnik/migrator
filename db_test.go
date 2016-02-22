package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListAllDBTenants(t *testing.T) {
	config, err := readConfigFromFile("test/migrator.yaml")
	assert.Nil(t, err)

	connector, err := CreateConnector(config)
	assert.Nil(t, err)

	connector.Init()
	defer connector.Dispose()

	tenants, err := connector.ListAllDBTenants()
	assert.Nil(t, err)

	assert.Len(t, tenants, 3)
	assert.Equal(t, []string{"abc", "def", "xyz"}, tenants)
}

func TestApplyMigrations(t *testing.T) {
	config, err := readConfigFromFile("test/migrator.yaml")
	assert.Nil(t, err)

	connector, err := CreateConnector(config)
	assert.Nil(t, err)

	connector.Init()
	defer connector.Dispose()

	allMigrations, _ := listAllMigrations(*config)
	dbMigrations, err := connector.ListAllDBMigrations()

	assert.Nil(t, err)
	assert.Len(t, dbMigrations, 0)

	migrationDefs := computeMigrationsToApply(allMigrations, dbMigrations)
	migrations, _ := loadMigrations(*config, migrationDefs)

	err = connector.ApplyMigrations(migrations)
	assert.Nil(t, err)

	dbMigrations, err = connector.ListAllDBMigrations()

	assert.Nil(t, err)
	assert.Len(t, dbMigrations, 12)
}
