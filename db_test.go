package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBCreateConnectorPanicUnknownDriver(t *testing.T) {
	config := &Config{}
	config.Driver = "abcxyz"

	assert.Panics(t, func() {
		CreateConnector(config)
	}, "Should panic because of unknown driver")
}

func TestListAllDBTenants(t *testing.T) {
	config := readConfigFromFile("test/migrator.yaml")

	connector := CreateConnector(config)

	connector.Init()
	defer connector.Dispose()

	tenants, err := connector.ListAllDBTenants()
	assert.Nil(t, err)

	assert.Len(t, tenants, 3)
	assert.Equal(t, []string{"abc", "def", "xyz"}, tenants)
}

func TestApplyMigrations(t *testing.T) {
	config := readConfigFromFile("test/migrator.yaml")

	connector := CreateConnector(config)

	connector.Init()
	defer connector.Dispose()

	allMigrations := listDiskMigrations(*config)
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
