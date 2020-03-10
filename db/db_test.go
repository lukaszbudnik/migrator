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

var (
	existingVersion     types.Version
	existingDBMigration types.DBMigration
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

func TestCreateVersion(t *testing.T) {
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

	results, version := connector.CreateVersion("commit-sha", types.ActionApply, false, migrationsToApply)

	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, "commit-sha", version.Name)
	assert.Equal(t, results.MigrationsGrandTotal+results.ScriptsGrandTotal, int32(len(version.DBMigrations)))
	assert.Equal(t, int32(noOfTenants), results.Tenants)
	assert.Equal(t, int32(3), results.SingleMigrations)
	assert.Equal(t, int32(2), results.SingleScripts)
	assert.Equal(t, int32(3), results.TenantMigrations)
	assert.Equal(t, int32(1), results.TenantScripts)
	assert.Equal(t, int32(noOfTenants*3), results.TenantMigrationsTotal)
	assert.Equal(t, int32(noOfTenants*1), results.TenantScriptsTotal)
	assert.Equal(t, int32(noOfTenants*3+3), results.MigrationsGrandTotal)
	assert.Equal(t, int32(noOfTenants*1+2), results.ScriptsGrandTotal)

	dbMigrationsAfter := connector.GetAppliedMigrations()
	lenAfter := len(dbMigrationsAfter)

	// 3 tenant migrations * no of tenants + 3 public
	// 1 tenant script * no of tenants + 2 public scripts
	expected := (3*noOfTenants + 3) + (1*noOfTenants + 2)
	assert.Equal(t, expected, lenAfter-lenBefore)
}

func TestCreateVersionEmptyMigrationArray(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	migrationsToApply := []types.Migration{}

	results, version := connector.CreateVersion("commit-sha", types.ActionApply, false, migrationsToApply)
	// empty migrations slice - no version created
	assert.Nil(t, version)
	assert.Equal(t, int32(0), results.MigrationsGrandTotal)
	assert.Equal(t, int32(0), results.ScriptsGrandTotal)
}

func TestCreateVersionDryRunMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	tn := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", tn), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", tn), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	tenants := sqlmock.NewRows([]string{"name"}).AddRow(tenant)
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	mock.ExpectPrepare("insert into migrator.migrator_versions").ExpectQuery().WithArgs("commit-sha")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into migrator.migrator_migrations").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	// get version
	rows := sqlmock.NewRows([]string{"vid", "vname", "vcreated", "mid", "name", "source_dir", "filename", "type", "db_schema", "created", "contents", "checksum"}).AddRow("123", "vname", time.Now(), "456", m.Name, m.SourceDir, m.File, m.MigrationType, tenant, time.Now(), m.Contents, m.CheckSum)
	mock.ExpectQuery("select").WillReturnRows(rows)
	// dry-run mode calls rollback instead of commit
	mock.ExpectRollback()

	// however the results contain correct dry-run data like number of applied migrations/scripts
	results, version := connector.CreateVersion("commit-sha", types.ActionApply, true, migrationsToApply)
	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, results.MigrationsGrandTotal+results.ScriptsGrandTotal, int32(len(version.DBMigrations)))
	assert.Equal(t, int32(1), results.MigrationsGrandTotal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionSyncMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	tn := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", tn), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", tn), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	tenants := sqlmock.NewRows([]string{"name"}).AddRow(tenant)
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	mock.ExpectPrepare("insert into migrator.migrator_versions").ExpectQuery().WithArgs("commit-sha")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	// get version
	rows := sqlmock.NewRows([]string{"vid", "vname", "vcreated", "mid", "name", "source_dir", "filename", "type", "db_schema", "created", "contents", "checksum"}).AddRow("123", "vname", time.Now(), "456", m.Name, m.SourceDir, m.File, m.MigrationType, tenant, time.Now(), m.Contents, m.CheckSum)
	mock.ExpectQuery("select").WillReturnRows(rows)
	mock.ExpectCommit()

	// sync the results contain correct data like number of applied migrations/scripts
	results, version := connector.CreateVersion("commit-sha", types.ActionSync, false, migrationsToApply)
	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, results.MigrationsGrandTotal+results.ScriptsGrandTotal, int32(len(version.DBMigrations)))
	assert.Equal(t, int32(1), results.MigrationsGrandTotal)

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

func TestCreateTenant(t *testing.T) {
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

	results, version := connector.CreateTenant("commit-sha", types.ActionApply, false, uniqueTenant, migrationsToApply)

	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, "commit-sha", version.Name)
	assert.Equal(t, results.MigrationsGrandTotal+results.ScriptsGrandTotal, int32(len(version.DBMigrations)))

	// applied only for one tenant - the newly added one
	assert.Equal(t, int32(1), results.Tenants)
	// just one tenant so total number of tenant migrations is equal to tenant migrations
	assert.Equal(t, int32(3), results.TenantMigrations)
	assert.Equal(t, int32(3), results.TenantMigrationsTotal)
}

func TestCreateTenantDryRunMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	tn := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", tn), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", tn), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	// tenant
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	mock.ExpectPrepare("insert into migrator.migrator_versions").ExpectQuery().WithArgs("commit-sha")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into migrator.migrator_migrations").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	// get version
	rows := sqlmock.NewRows([]string{"vid", "vname", "vcreated", "mid", "name", "source_dir", "filename", "type", "db_schema", "created", "contents", "checksum"}).AddRow("123", "vname", time.Now(), "456", m.Name, m.SourceDir, m.File, m.MigrationType, tenant, time.Now(), m.Contents, m.CheckSum)
	mock.ExpectQuery("select").WillReturnRows(rows)
	// dry-run mode calls rollback instead of commit
	mock.ExpectRollback()

	// however the results contain correct dry-run data like number of applied migrations/scripts
	results, version := connector.CreateTenant("commit-sha", types.ActionApply, true, tenant, migrationsToApply)
	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, results.MigrationsGrandTotal+results.ScriptsGrandTotal, int32(len(version.DBMigrations)))
	assert.Equal(t, int32(1), results.MigrationsGrandTotal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTenantSyncMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	tn := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", tn), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", tn), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	// tenant
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(0, 0))
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	mock.ExpectPrepare("insert into migrator.migrator_versions").ExpectQuery().WithArgs("commit-sha")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	// get version
	rows := sqlmock.NewRows([]string{"vid", "vname", "vcreated", "mid", "name", "source_dir", "filename", "type", "db_schema", "created", "contents", "checksum"}).AddRow("123", "vname", time.Now(), "456", m.Name, m.SourceDir, m.File, m.MigrationType, tenant, time.Now(), m.Contents, m.CheckSum)
	mock.ExpectQuery("select").WillReturnRows(rows)
	mock.ExpectCommit()

	// sync results contain correct data like number of applied migrations/scripts
	results, version := connector.CreateTenant("commit-sha", types.ActionSync, false, tenant, migrationsToApply)
	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, results.MigrationsGrandTotal+results.ScriptsGrandTotal, int32(len(version.DBMigrations)))
	assert.Equal(t, int32(1), results.MigrationsGrandTotal)

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

func TestGetVersions(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	versions := connector.GetVersions()

	assert.True(t, len(versions) >= 2)
	// versions are sorted from newest (highest ID) to oldest (lowest ID)
	assert.True(t, versions[0].ID > versions[1].ID)

	existingVersion = versions[0]
	existingDBMigration = existingVersion.DBMigrations[0]
}

func TestGetVersionsByFile(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	versions := connector.GetVersionsByFile(existingVersion.DBMigrations[0].File)
	version := versions[0]
	assert.Equal(t, existingVersion.ID, version.ID)
	assert.Equal(t, existingVersion.DBMigrations[0].File, version.DBMigrations[0].File)
	assert.True(t, len(version.DBMigrations) > 0)
}

func TestGetVersionByID(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	version, err := connector.GetVersionByID(existingVersion.ID)
	assert.Nil(t, err)
	assert.Equal(t, existingVersion.ID, version.ID)
	assert.True(t, len(version.DBMigrations) > 0)
}

func TestGetVersionByIDNotFound(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	version, err := connector.GetVersionByID(-1)
	assert.Nil(t, version)
	assert.Equal(t, "Version not found ID: -1", err.Error())
}

func TestGetDBMigrationByID(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	dbMigration, err := connector.GetDBMigrationByID(existingDBMigration.ID)
	assert.Nil(t, err)
	assert.Equal(t, existingDBMigration.ID, dbMigration.ID)
}

func TestGetDBMigrationByIDNotFound(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	dbMigration, err := connector.GetDBMigrationByID(-1)
	assert.Nil(t, dbMigration)
	assert.Equal(t, "DB migration not found ID: -1", err.Error())
}
