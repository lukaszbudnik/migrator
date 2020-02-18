package db

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func newTestContext() context.Context {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, common.RequestIDKey{}, time.Now().Nanosecond())
	return ctx
}

func TestDBCreateConnectorPanicUnknownDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "abcxyz"

	assert.PanicsWithValue(t, "Failed to create Connector unknown driver: abcxyz", func() {
		New(newTestContext(), config)
	})
}

func TestConnectorInitPanicConnectionError(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.DataSource = strings.Replace(config.DataSource, "127.0.0.1", "1.0.0.1", -1)

	didPanic := false
	var message interface{}
	func() {

		defer func() {
			if message = recover(); message != nil {
				didPanic = true
			}
		}()

		New(newTestContext(), config)

	}()
	assert.True(t, didPanic)
	assert.Contains(t, message, "Failed to connect to database")
}

func TestGetTenants(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	tenants := connector.GetTenants()

	assert.True(t, len(tenants) >= 3)
	assert.Contains(t, tenants, types.Tenant{Name: "abc"})
	assert.Contains(t, tenants, types.Tenant{Name: "def"})
	assert.Contains(t, tenants, types.Tenant{Name: "xyz"})
}

func TestApplyMigrations(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	tenants := connector.GetTenants()
	noOfTenants := len(tenants)

	dbMigrationsBefore := connector.GetAppliedMigrations()
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

	results := connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)

	assert.Equal(t, noOfTenants, results.Tenants)
	assert.Equal(t, 3, results.SingleMigrations)
	assert.Equal(t, 2, results.SingleScripts)
	assert.Equal(t, 3, results.TenantMigrations)
	assert.Equal(t, 1, results.TenantScripts)
	assert.Equal(t, noOfTenants*3, results.TenantMigrationsTotal)
	assert.Equal(t, noOfTenants*1, results.TenantScriptsTotal)
	assert.Equal(t, noOfTenants*3+3, results.MigrationsGrandTotal)
	assert.Equal(t, noOfTenants*1+2, results.ScriptsGrandTotal)

	dbMigrationsAfter := connector.GetAppliedMigrations()
	lenAfter := len(dbMigrationsAfter)

	// 3 tenant migrations * no of tenants + 3 public
	// 1 tenant script * no of tenants + 2 public scripts
	expected := (3*noOfTenants + 3) + (1*noOfTenants + 2)
	assert.Equal(t, expected, lenAfter-lenBefore)
}

func TestApplyMigrationsEmptyMigrationArray(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	migrationsToApply := []types.Migration{}

	results := connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)

	assert.Equal(t, 0, results.MigrationsGrandTotal)
	assert.Equal(t, 0, results.ScriptsGrandTotal)
}

func TestApplyMigrationsDryRunMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	tenants := sqlmock.NewRows([]string{"name"}).AddRow(tenant)
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnResult(sqlmock.NewResult(1, 1))

	// dry-run mode calls rollback instead of commit
	mock.ExpectRollback()

	// however the results contain correct dry-run data like number of applied migrations/scripts
	results := connector.ApplyMigrations(types.ModeTypeDryRun, migrationsToApply)

	assert.Equal(t, 1, results.MigrationsGrandTotal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyMigrationsSyncMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	tenants := sqlmock.NewRows([]string{"name"}).AddRow(tenant)
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// sync the results contain correct data like number of applied migrations/scripts
	results := connector.ApplyMigrations(types.ModeTypeSync, migrationsToApply)

	assert.Equal(t, 1, results.MigrationsGrandTotal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTenantsSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	tenantSelectSQL := connector.getTenantSelectSQL()

	assert.Equal(t, "select name from migrator.migrator_tenants", tenantSelectSQL)
}

func TestGetTenantsSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	tenantSelectSQL := connector.getTenantSelectSQL()

	assert.Equal(t, "select somename from someschema.sometable", tenantSelectSQL)
}

func TestGetSchemaPlaceHolderDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "{schema}", placeholder)
}

func TestGetSchemaPlaceHolderOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "[schema]", placeholder)
}

func TestAddTenantAndApplyMigrations(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()
	t3 := time.Now().UnixNano()

	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "drop table if exists {schema}.settings"}
	tenant2 := types.Migration{Name: fmt.Sprintf("%v.sql", t2), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t2), MigrationType: types.MigrationTypeTenantMigration, Contents: "create table {schema}.settings (k int, v text)"}
	tenant3 := types.Migration{Name: fmt.Sprintf("%v.sql", t3), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t3), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456')"}

	migrationsToApply := []types.Migration{tenant1, tenant2, tenant3}

	uniqueTenant := fmt.Sprintf("new_test_tenant_%v", time.Now().UnixNano())

	results := connector.AddTenantAndApplyMigrations(types.ModeTypeApply, uniqueTenant, migrationsToApply)

	// applied only for one tenant - the newly added one
	assert.Equal(t, 1, results.Tenants)
	// just one tenant so total number of tenant migrations is equal to tenant migrations
	assert.Equal(t, 3, results.TenantMigrations)
	assert.Equal(t, 3, results.TenantMigrationsTotal)
}

func TestAddTenantAndApplyMigrationsDryRunMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnResult(sqlmock.NewResult(1, 1))

	// dry-run mode calls rollback instead of commit
	mock.ExpectRollback()

	// however the results contain correct dry-run data like number of applied migrations/scripts
	results := connector.AddTenantAndApplyMigrations(types.ModeTypeDryRun, tenant, migrationsToApply)

	assert.Equal(t, 1, results.MigrationsGrandTotal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsSyncMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// sync results contain correct data like number of applied migrations/scripts
	results := connector.AddTenantAndApplyMigrations(types.ModeTypeSync, tenant, migrationsToApply)

	assert.Equal(t, 1, results.MigrationsGrandTotal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTenantInsertSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into someschema.sometable (somename) values ($1)", tenantInsertSQL)
}
