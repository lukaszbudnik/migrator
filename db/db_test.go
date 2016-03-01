package db

// These are integration tests which talk to database.
// These tests are almost self-contain
// they only depended on config package (reading config from file)

import (
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestDBCreateConnectorPanicUnknownDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "abcxyz"

	assert.Panics(t, func() {
		CreateConnector(config)
	}, "Should panic because of unknown driver")
}

func TestDBCreateConnectorPostgresDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "postgres"
	connector := CreateConnector(config)
	connectorName := reflect.TypeOf(connector).String()
	assert.Equal(t, "*db.postgreSQLConnector", connectorName)
}

func TestDBCreateConnectorMymysqlDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "mymysql"
	connector := CreateConnector(config)
	connectorName := reflect.TypeOf(connector).String()
	assert.Equal(t, "*db.mySQLConnector", connectorName)
}

func TestDBConnectorInitPanicConnectionError(t *testing.T) {
	config := config.FromFile("../test/migrator-test-non-existing-db.yaml")

	connector := CreateConnector(config)
	assert.Panics(t, func() {
		connector.Init()
	}, "Should panic because of connection error")
}

func TestDBGetTenantsPanicSQLSyntaxError(t *testing.T) {
	config := config.FromFile("../test/migrator.yaml")
	config.TenantsSQL = "sadfdsfdsf"
	connector := CreateConnector(config)
	connector.Init()
	assert.Panics(t, func() {
		connector.GetTenants()
	}, "Should panic because of tenants SQL syntax error")
}

func TestDBApplyMigrationsPanicSQLSyntaxError(t *testing.T) {
	config := config.FromFile("../test/migrator.yaml")
	config.SingleSchemas = []string{"error"}

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()
	m := types.MigrationDefinition{"201602220002.sql", "error", "error/201602220002.sql", types.MigrationTypeTenantSchema}
	ms := []types.Migration{{m, "createtablexyx ( idint primary key (id) )"}}

	assert.Panics(t, func() {
		connector.ApplyMigrations(ms)
	}, "Should panic because of migration SQL syntax error")
}

func TestDBGetTenants(t *testing.T) {
	config := config.FromFile("../test/migrator.yaml")

	connector := CreateConnector(config)

	connector.Init()
	defer connector.Dispose()

	tenants := connector.GetTenants()

	assert.Len(t, tenants, 3)
	assert.Equal(t, []string{"abc", "def", "xyz"}, tenants)
}

func TestDBApplyMigrations(t *testing.T) {
	config := config.FromFile("../test/migrator.yaml")

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()

	dbMigrationsBefore := connector.GetMigrations()
	lenBefore := len(dbMigrationsBefore)

	p1 := time.Now().UnixNano()
	p2 := time.Now().UnixNano()
	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()

	publicdef1 := types.MigrationDefinition{fmt.Sprintf("%v.sql", p1), "public", fmt.Sprintf("public/%v.sql", p1), types.MigrationTypeSingleSchema}
	publicdef2 := types.MigrationDefinition{fmt.Sprintf("%v.sql", p2), "public", fmt.Sprintf("public/%v.sql", p2), types.MigrationTypeSingleSchema}
	public1 := types.Migration{publicdef1, "create table if not exists {schema}.modules ( k int, v text )"}
	public2 := types.Migration{publicdef2, "insert into {schema}.modules values ( 123, '123' )"}

	tenantdef1 := types.MigrationDefinition{fmt.Sprintf("%v.sql", t1), "tenants", fmt.Sprintf("tenants/%v.sql", t1), types.MigrationTypeTenantSchema}
	tenantdef2 := types.MigrationDefinition{fmt.Sprintf("%v.sql", t2), "tenants", fmt.Sprintf("tenants/%v.sql", t2), types.MigrationTypeTenantSchema}
	tenant1 := types.Migration{tenantdef1, "create table if not exists {schema}.settings (k int, v text) "}
	tenant2 := types.Migration{tenantdef2, "insert into {schema}.settings values (456, '456') "}

	migrationsToApply := []types.Migration{public1, public2, tenant1, tenant2}

	connector.ApplyMigrations(migrationsToApply)

	dbMigrationsAfter := connector.GetMigrations()
	lenAfter := len(dbMigrationsAfter)

	assert.Equal(t, lenAfter-lenBefore, 8)
}
