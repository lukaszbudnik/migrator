package db

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func TestDoInitCannotBeginTransactionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "sqlmock"
	connector := baseConnector{config, nil, nil}

	mock.ExpectBegin().WillReturnError(errors.New("trouble maker"))

	err = connector.doInit(db)
	assert.Equal(t, "Could not start DB transaction: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDoInitCannotCreateMigratorSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnError(errors.New("trouble maker"))

	err = connector.doInit(db)
	assert.Equal(t, "Could not create migrator schema: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDoInitCannotCreateMigratorMigrationsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnError(errors.New("trouble maker"))

	err = connector.doInit(db)
	assert.Equal(t, "Could not create migrations table: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDoInitCannotCreateMigratorTenantsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	mock.ExpectBegin()
	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("create schema").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnRows()
	mock.ExpectQuery("create table").WillReturnError(errors.New("trouble maker"))

	err = connector.doInit(db)
	assert.Equal(t, "Could not create default tenants table: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBGetTenantsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	connector.db = db

	_, err = connector.GetTenants()
	assert.Equal(t, "Could not query tenants: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBGetMigrationsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	connector.db = db

	_, err = connector.GetDBMigrations()
	assert.Equal(t, "Could not query DB migrations: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyTransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	rows := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(rows)
	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	connector.db = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	_, err = connector.ApplyMigrations(newTestContext(), migrationsToApply)
	assert.NotNil(t, err)
	assert.Equal(t, "trouble maker tx.Begin()", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyInsertMigrationPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	_, err = connector.ApplyMigrations(newTestContext(), migrationsToApply)
	assert.Equal(t, "Could not create prepared statement: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyMigrationSQLError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	_, err = connector.ApplyMigrations(newTestContext(), migrationsToApply)
	assert.Equal(t, "SQL migration failed: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestApplyInsertMigrationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	tenants := sqlmock.NewRows([]string{"name"}).AddRow(tenant)
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	_, err = connector.ApplyMigrations(newTestContext(), migrationsToApply)
	assert.Equal(t, "Failed to add migration entry: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantTransactionBeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	connector.db = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	_, err = connector.AddTenantAndApplyMigrations(newTestContext(), "newtenant", migrationsToApply)
	assert.NotNil(t, err)
	assert.Equal(t, "trouble maker tx.Begin()", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsCreateSchemaError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	_, err = connector.AddTenantAndApplyMigrations(newTestContext(), "newtenant", migrationsToApply)
	assert.Equal(t, "Create schema failed, transaction rollback was called: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsInsertTenantPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	_, err = connector.AddTenantAndApplyMigrations(newTestContext(), "newtenant", migrationsToApply)
	assert.Equal(t, "Could not create prepared statement: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationsInsertTenantError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	tenant := "tenant"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	t1 := time.Now().UnixNano()
	m1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m1}

	_, err = connector.AddTenantAndApplyMigrations(newTestContext(), tenant, migrationsToApply)
	assert.Equal(t, "Failed to add tenant entry: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationInsertMigrationPreparedStatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	tenant := "tenant"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	t1 := time.Now().UnixNano()
	m1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m1}

	_, err = connector.AddTenantAndApplyMigrations(newTestContext(), tenant, migrationsToApply)
	assert.Equal(t, "Could not create prepared statement: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationMigrationSQLError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	tenant := "tenant"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	t1 := time.Now().UnixNano()
	m1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m1}

	_, err = connector.AddTenantAndApplyMigrations(newTestContext(), tenant, migrationsToApply)
	assert.Equal(t, "SQL migration failed: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddTenantAndApplyMigrationInsertMigrationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "postgres"
	dialect, err := newDialect(config)
	assert.Nil(t, err)
	connector := baseConnector{config, dialect, nil}

	tenant := "tenant"
	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantMigration, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.db = db

	_, err = connector.AddTenantAndApplyMigrations(newTestContext(), tenant, migrationsToApply)
	assert.Equal(t, "Failed to add migration entry: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
