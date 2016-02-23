package main

// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * disk_test.go
// * migrations_test.go

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestDBCreateConnectorPanicUnknownDriver(t *testing.T) {
	config := &Config{}
	config.Driver = "abcxyz"

	assert.Panics(t, func() {
		CreateConnector(config)
	}, "Should panic because of unknown driver")
}

func TestDBCreateConnectorPostgresDriver(t *testing.T) {
	config := &Config{}
	config.Driver = "postgres"
	connector := CreateConnector(config)
	connectorName := reflect.TypeOf(connector).String()
	assert.Equal(t, "*main.postgresqlConnector", connectorName)
}

func TestDBCreateConnectorMymysqlDriver(t *testing.T) {
	config := &Config{}
	config.Driver = "mymysql"
	connector := CreateConnector(config)
	connectorName := reflect.TypeOf(connector).String()
	assert.Equal(t, "*main.mysqlConnector", connectorName)
}

func TestDBConnectorInitPanicConnectionError(t *testing.T) {
	config := readConfigFromFile("test/migrator-test-non-existing-db.yaml")

	connector := CreateConnector(config)
	assert.Panics(t, func() {
		connector.Init()
	}, "Should panic because of connection error")
}

func TestDBGetTenantsPanicSQLSyntaxError(t *testing.T) {
	config := readConfigFromFile("test/migrator.yaml")
	config.TenantsSQL = "sadfdsfdsf"
	connector := CreateConnector(config)
	connector.Init()
	assert.Panics(t, func() {
		connector.GetDBTenants()
	}, "Should panic because of tenants SQL syntax error")
}

func TestDBApplyMigrationsPanicSQLSyntaxError(t *testing.T) {
	config := readConfigFromFile("test/migrator.yaml")
	config.SingleSchemas = []string{"error"}

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()
	m := MigrationDefinition{"201602220002.sql", "tenants", "tenants/201602220002.sql", ModeTenantSchema}
	ms := []Migration{{m, "createtablexyx ( idint primary key (id) )"}}

	assert.Panics(t, func() {
		connector.ApplyMigrations(ms)
	}, "Should panic because of migration SQL syntax error")
}

func TestDBGetTenants(t *testing.T) {
	config := readConfigFromFile("test/migrator.yaml")

	connector := CreateConnector(config)

	connector.Init()
	defer connector.Dispose()

	tenants := connector.GetDBTenants()

	assert.Len(t, tenants, 3)
	assert.Equal(t, []string{"abc", "def", "xyz"}, tenants)
}

func TestDBApplyMigrations(t *testing.T) {
	config := readConfigFromFile("test/migrator.yaml")

	loader := CreateLoader(config)

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()

	dbMigrationsBefore := connector.GetDBMigrations()

	diskMigrations := loader.GetDiskMigrations()
	migrationsToApply := computeMigrationsToApply(diskMigrations, dbMigrationsBefore)

	connector.ApplyMigrations(migrationsToApply)

	dbMigrationsAfter := connector.GetDBMigrations()

	assert.Len(t, dbMigrationsBefore, 0)
	assert.Len(t, dbMigrationsAfter, 12)
}
