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

func TestDBCreateConnectorMysqlDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "mysql"
	connector := CreateConnector(config)
	connectorName := reflect.TypeOf(connector).String()
	assert.Equal(t, "*db.mySQLConnector", connectorName)
}

func TestDBConnectorInitPanicConnectionError(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.DataSource = ""

	connector := CreateConnector(config)
	assert.Panics(t, func() {
		connector.Init()
	}, "Should panic because of connection error")
}

func TestDBGetTenantsPanicSQLSyntaxError(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.TenantSelectSql = "sadfdsfdsf"
	connector := CreateConnector(config)
	connector.Init()
	assert.Panics(t, func() {
		connector.GetTenants()
	}, "Should panic because of tenants SQL syntax error")
}

func TestDBApplyMigrationsPanicSQLSyntaxError(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

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
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)

	connector.Init()
	defer connector.Dispose()

	tenants := connector.GetTenants()

	// todo more than 3 and contains abc, def, xyz
	assert.True(t, len(tenants) >= 3)
	assert.Contains(t, tenants, "abc")
	assert.Contains(t, tenants, "def")
	assert.Contains(t, tenants, "xyz")
}

func TestDBApplyMigrations(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()

	tenants := connector.GetTenants()
	lenTenants := len(tenants)

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

	// 2 tenant migrations * no of tenants + 2 public
	expected := lenTenants*2 + 2
	assert.Equal(t, expected, lenAfter-lenBefore)
}

func TestDBApplyMigrationsEmptyMigrationArray(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()

	dbMigrationsBefore := connector.GetMigrations()
	lenBefore := len(dbMigrationsBefore)

	migrationsToApply := []types.Migration{}

	connector.ApplyMigrations(migrationsToApply)

	dbMigrationsAfter := connector.GetMigrations()
	lenAfter := len(dbMigrationsAfter)

	assert.Equal(t, lenAfter, lenBefore)
}

func TestGetTenantsSqlDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantSelectSql := connector.GetTenantSelectSql()

	assert.Equal(t, "select name from public.migrator_tenants", tenantSelectSql)
}

func TestGetTenantsSqlOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantSelectSql := connector.GetTenantSelectSql()

	assert.Equal(t, "select somename from someschema.sometable", tenantSelectSql)
}

func TestGetSchemaPlaceHolderDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	placeholder := connector.GetSchemaPlaceHolder()

	assert.Equal(t, "{schema}", placeholder)
}

func TestGetSchemaPlaceHolderOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	placeholder := connector.GetSchemaPlaceHolder()

	assert.Equal(t, "[schema]", placeholder)
}

func TestAddTenantAndApplyMigrations(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()

	dbMigrationsBefore := connector.GetMigrations()
	lenBefore := len(dbMigrationsBefore)

	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()
	t3 := time.Now().UnixNano()

	tenantdef1 := types.MigrationDefinition{fmt.Sprintf("%v.sql", t1), "tenants", fmt.Sprintf("tenants/%v.sql", t1), types.MigrationTypeTenantSchema}
	tenantdef2 := types.MigrationDefinition{fmt.Sprintf("%v.sql", t2), "tenants", fmt.Sprintf("tenants/%v.sql", t2), types.MigrationTypeTenantSchema}
	tenantdef3 := types.MigrationDefinition{fmt.Sprintf("%v.sql", t3), "tenants", fmt.Sprintf("tenants/%v.sql", t3), types.MigrationTypeTenantSchema}
	tenant1 := types.Migration{tenantdef1, "create schema {schema}"}
	tenant2 := types.Migration{tenantdef2, "create table if not exists {schema}.settings (k int, v text) "}
	tenant3 := types.Migration{tenantdef3, "insert into {schema}.settings values (456, '456') "}

	migrationsToApply := []types.Migration{tenant1, tenant2, tenant3}

	unique_tenant := fmt.Sprintf("new_test_tenant_%v", time.Now().UnixNano())

	connector.AddTenantAndApplyMigrations(unique_tenant, migrationsToApply)

	dbMigrationsAfter := connector.GetMigrations()
	lenAfter := len(dbMigrationsAfter)

	assert.Equal(t, 3, lenAfter-lenBefore)
}

func TestMySQLGetMigrationInsertSql(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"

	dialect := CreateDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSql()

	assert.Equal(t, "insert into public.migrator_migrations (name, source_dir, file, type, db_schema) values (?, ?, ?, ?, ?)", insertMigrationSQL)
}

func TestPostgreSQLGetMigrationInsertSql(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := CreateDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSql()

	assert.Equal(t, "insert into public.migrator_migrations (name, source_dir, file, type, db_schema) values ($1, $2, $3, $4, $5)", insertMigrationSQL)
}

func TestMySQLGetTenantInsertSqlDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantInsertSql := connector.GetTenantInsertSql()

	assert.Equal(t, "insert into public.migrator_tenants (name) values (?)", tenantInsertSql)
}

func TestPostgreSQLGetTenantInsertSqlDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantInsertSql := connector.GetTenantInsertSql()

	assert.Equal(t, "insert into public.migrator_tenants (name) values ($1)", tenantInsertSql)
}

func TestGetTenantInsertSqlOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantInsertSql := connector.GetTenantInsertSql()

	assert.Equal(t, "insert into XXX", tenantInsertSql)
}
