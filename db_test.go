package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListAllDBTenants(t *testing.T) {
	config, err := readConfigFromFile("test/migrator.yaml")
	db, err := sql.Open(config.Driver, config.DataSource)

	tenants, err := listAllDBTenants(*config, db)

	assert.Nil(t, err)
	assert.Len(t, tenants, 3)
	assert.Equal(t, []string{"abc", "def", "xyz"}, tenants)

	db.Close()
}

func TestApplyMigrations(t *testing.T) {
	config, _ := readConfigFromFile("test/migrator.yaml")

	allMigrations, _ := listAllMigrations(*config)
	dbMigrations, err := listAllDBMigrations(*config)

	assert.Nil(t, err)
	assert.Len(t, dbMigrations, 0)

	migrationDefs := computeMigrationsToApply(allMigrations, dbMigrations)
	migrations, _ := loadMigrations(*config, migrationDefs)

	err = applyMigrations(*config, migrations)

	assert.Nil(t, err)

	dbMigrations, err = listAllDBMigrations(*config)

	assert.Nil(t, err)
	assert.Len(t, dbMigrations, 6)
}
