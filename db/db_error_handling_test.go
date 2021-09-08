package db

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func TestInitCannotBeginTransactionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "sqlmock"
	connector := baseConnector{newTestContext(), config, nil, db, false}

	mock.ExpectBegin().WillReturnError(errors.New("trouble maker"))

	initErr := connector.init()

	assert.NotNil(t, initErr)
	assert.Contains(t, initErr.Error(), "could not start DB transaction: trouble maker")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInitCannotCreateMigratorSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, false}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectExec("create schema").WillReturnError(errors.New("trouble maker"))

	initErr := connector.init()

	assert.NotNil(t, initErr)
	assert.Contains(t, initErr.Error(), "could not create migrator schema: trouble maker")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInitCannotCreateMigratorMigrationsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, false}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("create table").WillReturnError(errors.New("trouble maker"))

	initErr := connector.init()

	assert.NotNil(t, initErr)
	assert.Contains(t, initErr.Error(), "could not create migrations table: trouble maker")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInitCannotCreateMigratorVersionsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, false}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("create table").WillReturnResult(sqlmock.NewResult(0, 0))
	// create versions table is a script
	mock.ExpectExec("begin").WillReturnError(errors.New("trouble maker"))

	initErr := connector.init()

	assert.NotNil(t, initErr)
	assert.Contains(t, initErr.Error(), "could not create versions table: trouble maker")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInitCannotCreateMigratorTenantsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, false}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("create table").WillReturnResult(sqlmock.NewResult(0, 0))
	// create versions table is a script
	mock.ExpectExec("begin").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("create table").WillReturnError(errors.New("trouble maker"))

	initErr := connector.init()

	assert.NotNil(t, initErr)
	assert.Contains(t, initErr.Error(), "could not create default tenants table: trouble maker")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInitCannotCommitTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, false}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("create table").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("begin").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("create table").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit().WillReturnError(errors.New("trouble maker"))

	initErr := connector.init()

	assert.NotNil(t, initErr)
	assert.Contains(t, initErr.Error(), "could not commit transaction: trouble maker")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTenantsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not query tenants: trouble maker", func() {
		connector.GetTenants()
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetAppliedMigrationsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not query DB migrations: trouble maker", func() {
		connector.GetAppliedMigrations()
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionTransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	rows := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(rows)
	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not start transaction: trouble maker tx.Begin()", func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionInsertVersionPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not create prepared statement for version: trouble maker", func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionInsertMigrationPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	mock.ExpectPrepare("insert into migrator.migrator_versions").ExpectQuery().WithArgs("commit-sha")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not create prepared statement for migration: trouble maker", func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionMigrationSQLError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	mock.ExpectPrepare("insert into migrator.migrator_versions").ExpectQuery().WithArgs("commit-sha")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectExec("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, fmt.Sprintf("SQL migration %v failed with error: trouble maker", tenant1.File), func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionInsertMigrationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
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
	mock.ExpectPrepare("insert into migrator.migrator_migrations").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	assert.PanicsWithValue(t, "Failed to add migration entry: trouble maker", func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionGetVersionError(t *testing.T) {
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
	mock.ExpectQuery("select").WillReturnError(errors.New("get version trouble maker"))

	assert.PanicsWithValue(t, "Could not query versions: get version trouble maker", func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionVersionNotFound(t *testing.T) {
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
	rows := sqlmock.NewRows([]string{"vid", "vname", "vcreated", "mid", "name", "source_dir", "filename", "type", "db_schema", "created", "contents", "checksum"})
	mock.ExpectQuery("select").WillReturnRows(rows)

	assert.PanicsWithValue(t, "Version not found ID: 0", func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateVersionMigrationsCommitError(t *testing.T) {
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
	mock.ExpectCommit().WillReturnError(errors.New("tx trouble maker"))

	assert.PanicsWithValue(t, "Could not commit transaction: tx trouble maker", func() {
		connector.CreateVersion("commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTenantTransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not start transaction: trouble maker tx.Begin()", func() {
		connector.CreateTenant("newtenant", "commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTenantCreateSchemaError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Create schema failed: trouble maker", func() {
		connector.CreateTenant("newtenant", "commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTenantInsertTenantPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not create prepared statement: trouble maker", func() {
		connector.CreateTenant("newtenant", "commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTenantInsertTenantError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	tenant := "tenant"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	m1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m1}

	assert.PanicsWithValue(t, "Failed to add tenant entry: trouble maker", func() {
		connector.CreateTenant(tenant, "commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTenantCommitError(t *testing.T) {
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
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	// get version
	rows := sqlmock.NewRows([]string{"vid", "vname", "vcreated", "mid", "name", "source_dir", "filename", "type", "db_schema", "created", "contents", "checksum"}).AddRow("123", "vname", time.Now(), "456", m.Name, m.SourceDir, m.File, m.MigrationType, tenant, time.Now(), m.Contents, m.CheckSum)
	mock.ExpectQuery("select").WillReturnRows(rows)
	mock.ExpectCommit().WillReturnError(errors.New("tx trouble maker"))

	assert.PanicsWithValue(t, "Could not commit transaction: tx trouble maker", func() {
		connector.CreateTenant(tenant, "commit-sha", types.ActionApply, migrationsToApply, false)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetVersionsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not query versions: trouble maker", func() {
		connector.GetVersions()
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetVersionsByFileError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not query versions: trouble maker", func() {
		connector.GetVersionsByFile("file")
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetVersionsByIDError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not query versions: trouble maker", func() {
		connector.GetVersionByID(0)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetDBMigrationByIDError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db, true}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not query DB migrations: trouble maker", func() {
		connector.GetDBMigrationByID(0)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
