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
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
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

		db := New(newTestContext(), config)
		db.GetTenants()

	}()
	assert.True(t, didPanic)
	assert.Contains(t, message, "Error initialising migrator: failed to connect to database")
}

func TestCreateVersionDryRunMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

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
	results, version := connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, true)
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
	connector := baseConnector{newTestContext(), config, dialect, db, true}

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
	results, version := connector.CreateVersion("commit-sha", types.ActionSync, migrationsToApply, false)
	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, results.MigrationsGrandTotal+results.ScriptsGrandTotal, int32(len(version.DBMigrations)))
	assert.Equal(t, int32(1), results.MigrationsGrandTotal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTenantsSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil, false}
	defer connector.Dispose()

	tenantSelectSQL := connector.getTenantSelectSQL()

	assert.Equal(t, "select somename from someschema.sometable", tenantSelectSQL)
}

func TestGetSchemaPlaceHolderDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator-postgresql.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil, false}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "{schema}", placeholder)
}

func TestGetSchemaPlaceHolderOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil, false}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "[schema]", placeholder)
}

func TestCreateTenantDryRunMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

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
	results, version := connector.CreateTenant(tenant, "commit-sha", types.ActionApply, migrationsToApply, true)
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
	connector := baseConnector{newTestContext(), config, dialect, db, true}

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
	results, version := connector.CreateTenant(tenant, "commit-sha", types.ActionSync, migrationsToApply, false)
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
	connector := baseConnector{newTestContext(), config, dialect, nil, false}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into someschema.sometable (somename) values ($1)", tenantInsertSQL)
}
