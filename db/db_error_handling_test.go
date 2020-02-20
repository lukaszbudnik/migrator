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
	connector := baseConnector{newTestContext(), config, nil, db}

	mock.ExpectBegin().WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not start DB transaction: trouble maker", func() {
		connector.init()
	})

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
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not create migrator schema: trouble maker", func() {
		connector.init()
	})

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
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not create migrations table: trouble maker", func() {
		connector.init()
	})

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
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnRows()
	// create versions table is a script
	mock.ExpectQuery("begin").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not create versions table: trouble maker", func() {
		connector.init()
	})

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
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnRows()
	// create versions table is a script
	mock.ExpectQuery("begin").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not create default tenants table: trouble maker", func() {
		connector.init()
	})

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
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnRows()
	mock.ExpectQuery("begin").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnRows()
	mock.ExpectCommit().WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not commit transaction: trouble maker", func() {
		connector.init()
	})

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
	connector := baseConnector{newTestContext(), config, dialect, db}

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
	connector := baseConnector{newTestContext(), config, dialect, db}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	assert.PanicsWithValue(t, "Could not query DB migrations: trouble maker", func() {
		connector.GetAppliedMigrations()
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyTransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	rows := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(rows)
	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not start transaction: trouble maker tx.Begin()", func() {
		connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyInsertVersionPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not create prepared statement for version: trouble maker", func() {
		connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyInsertMigrationPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not create prepared statement for migration: trouble maker", func() {
		connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyMigrationSQLError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectExec("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, fmt.Sprintf("SQL migration %v failed with error: trouble maker", tenant1.File), func() {
		connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyInsertMigrationError(t *testing.T) {
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
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	assert.PanicsWithValue(t, "Failed to add migration entry: trouble maker", func() {
		connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyMigrationsCommitError(t *testing.T) {
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
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit().WillReturnError(errors.New("tx trouble maker"))

	assert.PanicsWithValue(t, "Could not commit transaction: tx trouble maker", func() {
		connector.ApplyMigrations(types.ModeTypeApply, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantTransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not start transaction: trouble maker tx.Begin()", func() {
		connector.AddTenantAndApplyMigrations(types.ModeTypeApply, "newtenant", migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsCreateSchemaError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Create schema failed: trouble maker", func() {
		connector.AddTenantAndApplyMigrations(types.ModeTypeApply, "newtenant", migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsInsertTenantPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	assert.PanicsWithValue(t, "Could not create prepared statement: trouble maker", func() {
		connector.AddTenantAndApplyMigrations(types.ModeTypeApply, "newtenant", migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsInsertTenantError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, db}

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
		connector.AddTenantAndApplyMigrations(types.ModeTypeApply, tenant, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsCommitError(t *testing.T) {
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
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(0, 0))
	// tenant
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(0, 0))
	// version
	mock.ExpectPrepare("insert into migrator.migrator_versions")
	// migration
	mock.ExpectPrepare("insert into migrator.migrator_migrations")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum, 0).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit().WillReturnError(errors.New("tx trouble maker"))

	assert.PanicsWithValue(t, "Could not commit transaction: tx trouble maker", func() {
		connector.AddTenantAndApplyMigrations(types.ModeTypeApply, tenant, migrationsToApply)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
