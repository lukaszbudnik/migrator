package db

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func newTestContext() context.Context {
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)

	ctx := context.TODO()
	ctx = context.WithValue(ctx, common.RequestIDKey{}, "123")
	ctx = context.WithValue(ctx, common.ActionKey{}, strings.Replace(details.Name(), "github.com/lukaszbudnik/migrator/db.", "", -1))
	return ctx
}

func TestDBCreateConnectorPanicUnknownDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "abcxyz"

	_, err := NewConnector(config)
	assert.Contains(t, err.Error(), "unknown driver")
}

func TestBaseConnectorPanicUnknownDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "sfsdf"
	connector := baseConnector{config, nil, nil}
	err := connector.Init()
	assert.Contains(t, err.Error(), "unknown driver")
}

func TestDBConnectorInitPanicConnectionError(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.DataSource = strings.Replace(config.DataSource, "127.0.0.1", "1.0.0.1", -1)

	connector, err := NewConnector(config)
	assert.Nil(t, err)
	err = connector.Init()
	assert.Contains(t, err.Error(), "Failed to connect to database")
}

func TestDBGetTenants(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector, err := NewConnector(config)
	assert.Nil(t, err)

	err = connector.Init()
	assert.Nil(t, err)
	defer connector.Dispose()

	tenants, err := connector.GetTenants()

	assert.Nil(t, err)
	assert.True(t, len(tenants) >= 3)
	assert.Contains(t, tenants, "abc")
	assert.Contains(t, tenants, "def")
	assert.Contains(t, tenants, "xyz")
}

func TestDBApplyMigrations(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector, err := NewConnector(config)
	assert.Nil(t, err)
	connector.Init()
	defer connector.Dispose()

	tenants, err := connector.GetTenants()
	assert.Nil(t, err)

	noOfTenants := len(tenants)

	dbMigrationsBefore, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenBefore := len(dbMigrationsBefore)

	p1 := time.Now().UnixNano()
	p2 := time.Now().UnixNano()
	p3 := time.Now().UnixNano()
	p4 := time.Now().UnixNano()
	p5 := time.Now().UnixNano()
	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()
	t3 := time.Now().UnixNano()
	t4 := time.Now().UnixNano()

	// public migrations
	public1 := types.Migration{Name: fmt.Sprintf("%v.sql", p1), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p1), MigrationType: types.MigrationTypeSingleMigration, Contents: "drop table if exists modules"}
	public2 := types.Migration{Name: fmt.Sprintf("%v.sql", p2), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p2), MigrationType: types.MigrationTypeSingleMigration, Contents: "create table modules ( k int, v text )"}
	public3 := types.Migration{Name: fmt.Sprintf("%v.sql", p3), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p3), MigrationType: types.MigrationTypeSingleMigration, Contents: "insert into modules values ( 123, '123' )"}

	// public scripts
	public4 := types.Migration{Name: fmt.Sprintf("%v.sql", p4), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p4), MigrationType: types.MigrationTypeSingleScript, Contents: "insert into modules values ( 1234, '1234' )"}
	public5 := types.Migration{Name: fmt.Sprintf("%v.sql", p5), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p5), MigrationType: types.MigrationTypeSingleScript, Contents: "insert into modules values ( 12345, '12345' )"}

	// tenant migrations
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "drop table if exists {schema}.settings"}
	tenant2 := types.Migration{Name: fmt.Sprintf("%v.sql", t2), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t2), MigrationType: types.MigrationTypeTenantMigration, Contents: "create table {schema}.settings (k int, v text)"}
	tenant3 := types.Migration{Name: fmt.Sprintf("%v.sql", t3), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t3), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}

	// tenant scripts
	tenant4 := types.Migration{Name: fmt.Sprintf("%v.sql", t4), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t4), MigrationType: types.MigrationTypeTenantScript, Contents: "insert into {schema}.settings values (456, '456') "}

	migrationsToApply := []types.Migration{public1, public2, public3, tenant1, tenant2, tenant3, public4, public5, tenant4}

	results, err := connector.ApplyMigrations(newTestContext(), migrationsToApply)
	assert.Nil(t, err)
	assert.Equal(t, noOfTenants, results.Tenants)
	assert.Equal(t, 3, results.SingleMigrations)
	assert.Equal(t, 2, results.SingleScripts)
	assert.Equal(t, 3, results.TenantMigrations)
	assert.Equal(t, 1, results.TenantScripts)
	assert.Equal(t, noOfTenants*3, results.TenantMigrationsTotal)
	assert.Equal(t, noOfTenants*1, results.TenantScriptsTotal)
	assert.Equal(t, noOfTenants*3+3, results.MigrationsTotal)
	assert.Equal(t, noOfTenants*1+2, results.ScriptsTotal)

	dbMigrationsAfter, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenAfter := len(dbMigrationsAfter)

	// 3 tenant migrations * no of tenants + 3 public
	// 1 tenant script * no of tenants + 2 public scripts
	expected := (3*noOfTenants + 3) + (1*noOfTenants + 2)
	assert.Equal(t, expected, lenAfter-lenBefore)
}

func TestDBApplyMigrationsEmptyMigrationArray(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector, err := NewConnector(config)
	assert.Nil(t, err)
	connector.Init()
	defer connector.Dispose()

	migrationsToApply := []types.Migration{}

	results, err := connector.ApplyMigrations(newTestContext(), migrationsToApply)
	assert.Nil(t, err)

	assert.Equal(t, 0, results.MigrationsTotal)
	assert.Equal(t, 0, results.ScriptsTotal)
}

func TestGetTenantsSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}
	defer connector.Dispose()

	tenantSelectSQL := connector.getTenantSelectSQL()

	assert.Equal(t, "select name from migrator.migrator_tenants", tenantSelectSQL)
}

func TestGetTenantsSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}
	defer connector.Dispose()

	tenantSelectSQL := connector.getTenantSelectSQL()

	assert.Equal(t, "select somename from someschema.sometable", tenantSelectSQL)
}

func TestGetSchemaPlaceHolderDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "{schema}", placeholder)
}

func TestGetSchemaPlaceHolderOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "[schema]", placeholder)
}

func TestAddTenantAndApplyMigrations(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}
	connector.Init()
	defer connector.Dispose()

	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()
	t3 := time.Now().UnixNano()

	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "drop table if exists {schema}.settings"}
	tenant2 := types.Migration{Name: fmt.Sprintf("%v.sql", t2), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t2), MigrationType: types.MigrationTypeTenantMigration, Contents: "create table {schema}.settings (k int, v text)"}
	tenant3 := types.Migration{Name: fmt.Sprintf("%v.sql", t3), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t3), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456')"}

	migrationsToApply := []types.Migration{tenant1, tenant2, tenant3}

	uniqueTenant := fmt.Sprintf("new_test_tenant_%v", time.Now().UnixNano())

	results, err := connector.AddTenantAndApplyMigrations(newTestContext(), uniqueTenant, migrationsToApply)
	assert.Nil(t, err)

	// applied only for one tenant - the newly added one
	assert.Equal(t, 1, results.Tenants)
	// just one tenant so total number of tenant migrations is equal to tenant migrations
	assert.Equal(t, 3, results.TenantMigrations)
	assert.Equal(t, 3, results.TenantMigrationsTotal)
}

func TestGetTenantInsertSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into someschema.sometable (somename) values ($1)", tenantInsertSQL)
}
