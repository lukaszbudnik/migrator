package db

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func TestDBCreateConnectorPanicUnknownDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "abcxyz"

	assert.Panics(t, func() {
		CreateConnector(config)
	}, "Should panic because of unknown driver")
}

func TestBaseConnectorPanicUnknownDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "sfsdf"
	connector := BaseConnector{config, nil, nil}
	err := connector.Init()
	assert.Contains(t, err.Error(), "sql: unknown driver")
}

func TestDBCreateDialectPostgreSQLDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "postgres"
	dialect := CreateDialect(config)
	assert.IsType(t, &postgreSQLDialect{}, dialect)
}

func TestDBCreateDialectMysqlDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "mysql"
	dialect := CreateDialect(config)
	assert.IsType(t, &mySQLDialect{}, dialect)
}

func TestDBCreateDialectMSSQLDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "sqlserver"
	dialect := CreateDialect(config)
	assert.IsType(t, &msSQLDialect{}, dialect)
}

func TestDBConnectorInitPanicConnectionError(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.DataSource = strings.Replace(config.DataSource, "127.0.0.1", "1.0.0.1", -1)

	connector := CreateConnector(config)
	err = connector.Init()
	assert.Contains(t, err.Error(), "Failed to connect to database")
}

func TestDBGetTenants(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)

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

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()

	tenants, err := connector.GetTenants()
	assert.Nil(t, err)

	lenTenants := len(tenants)

	dbMigrationsBefore, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenBefore := len(dbMigrationsBefore)

	p1 := time.Now().UnixNano()
	p2 := time.Now().UnixNano()
	p3 := time.Now().UnixNano()
	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()
	t3 := time.Now().UnixNano()

	public1 := types.Migration{Name: fmt.Sprintf("%v.sql", p1), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p1), MigrationType: types.MigrationTypeSingleSchema, Contents: "drop table if exists modules"}
	public2 := types.Migration{Name: fmt.Sprintf("%v.sql", p2), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p2), MigrationType: types.MigrationTypeSingleSchema, Contents: "create table modules ( k int, v text )"}
	public3 := types.Migration{Name: fmt.Sprintf("%v.sql", p3), SourceDir: "public", File: fmt.Sprintf("public/%v.sql", p3), MigrationType: types.MigrationTypeSingleSchema, Contents: "insert into modules values ( 123, '123' )"}

	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "drop table if exists {schema}.settings"}
	tenant2 := types.Migration{Name: fmt.Sprintf("%v.sql", t2), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t2), MigrationType: types.MigrationTypeTenantSchema, Contents: "create table {schema}.settings (k int, v text)"}
	tenant3 := types.Migration{Name: fmt.Sprintf("%v.sql", t3), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t2), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}

	migrationsToApply := []types.Migration{public1, public2, public3, tenant1, tenant2, tenant3}

	connector.ApplyMigrations(migrationsToApply)

	dbMigrationsAfter, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenAfter := len(dbMigrationsAfter)

	// 3 tenant migrations * no of tenants + 3 public
	expected := lenTenants*3 + 3
	assert.Equal(t, expected, lenAfter-lenBefore)
}

func TestDBApplyMigrationsEmptyMigrationArray(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	connector.Init()
	defer connector.Dispose()

	dbMigrationsBefore, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenBefore := len(dbMigrationsBefore)

	migrationsToApply := []types.Migration{}

	connector.ApplyMigrations(migrationsToApply)

	dbMigrationsAfter, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenAfter := len(dbMigrationsAfter)

	assert.Equal(t, lenAfter, lenBefore)
}

func TestGetTenantsSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	tenantSelectSQL := connector.getTenantSelectSQL()

	assert.Equal(t, "select name from migrator.migrator_tenants", tenantSelectSQL)
}

func TestGetTenantsSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	tenantSelectSQL := connector.getTenantSelectSQL()

	assert.Equal(t, "select somename from someschema.sometable", tenantSelectSQL)
}

func TestGetSchemaPlaceHolderDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "{schema}", placeholder)
}

func TestGetSchemaPlaceHolderOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	placeholder := connector.getSchemaPlaceHolder()

	assert.Equal(t, "[schema]", placeholder)
}

func TestAddTenantAndApplyMigrations(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := BaseConnector{config, CreateDialect(config), nil}
	connector.Init()
	defer connector.Dispose()

	dbMigrationsBefore, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenBefore := len(dbMigrationsBefore)

	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()
	t3 := time.Now().UnixNano()

	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "drop table if exists {schema}.settings"}
	tenant2 := types.Migration{Name: fmt.Sprintf("%v.sql", t2), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t2), MigrationType: types.MigrationTypeTenantSchema, Contents: "create table {schema}.settings (k int, v text)"}
	tenant3 := types.Migration{Name: fmt.Sprintf("%v.sql", t3), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t3), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456')"}

	migrationsToApply := []types.Migration{tenant1, tenant2, tenant3}

	uniqueTenant := fmt.Sprintf("new_test_tenant_%v", time.Now().UnixNano())

	connector.AddTenantAndApplyMigrations(uniqueTenant, migrationsToApply)

	dbMigrationsAfter, err := connector.GetDBMigrations()
	assert.Nil(t, err)

	lenAfter := len(dbMigrationsAfter)

	assert.Equal(t, 3, lenAfter-lenBefore)
}

func TestMySQLGetMigrationInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"

	dialect := CreateDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_migrations (name, source_dir, filename, type, db_schema, contents, checksum) values (?, ?, ?, ?, ?, ?, ?)", insertMigrationSQL)
}

func TestPostgreSQLGetMigrationInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := CreateDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_migrations (name, source_dir, filename, type, db_schema, contents, checksum) values ($1, $2, $3, $4, $5, $6, $7)", insertMigrationSQL)
}

func TestMSSQLGetMigrationInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := CreateDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_migrations (name, source_dir, filename, type, db_schema, contents, checksum) values (@p1, @p2, @p3, @p4, @p5, @p6, @p7)", insertMigrationSQL)
}

func TestMySQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values (?)", tenantInsertSQL)
}

func TestPostgreSQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values ($1)", tenantInsertSQL)
}

func TestMSSQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values (@p1)", tenantInsertSQL)
}

func TestGetTenantInsertSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := BaseConnector{config, CreateDialect(config), nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into someschema.sometable (somename) values ($1)", tenantInsertSQL)
}

func TestMSSQLDialectGetCreateTenantsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := CreateDialect(config)

	createTenantsTableSQL := dialect.GetCreateTenantsTableSQL()

	expected := `
IF NOT EXISTS (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_tenants')
BEGIN
  create table [migrator].migrator_tenants (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP
  );
END
`

	assert.Equal(t, expected, createTenantsTableSQL)
}

func TestMSSQLDialectGetCreateMigrationsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := CreateDialect(config)

	createMigrationsTableSQL := dialect.GetCreateMigrationsTableSQL()

	expected := `
IF NOT EXISTS (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_migrations')
BEGIN
  create table [migrator].migrator_migrations (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    source_dir varchar(200) not null,
    filename varchar(200) not null,
    type int not null,
    db_schema varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP,
		contents text,
		checksum varchar(64)
  );
END
`

	assert.Equal(t, expected, createMigrationsTableSQL)
}

func TestBaseDialectGetCreateTenantsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := CreateDialect(config)

	createTenantsTableSQL := dialect.GetCreateTenantsTableSQL()

	expected := `
create table if not exists migrator.migrator_tenants (
  id serial primary key,
  name varchar(200) not null,
  created timestamp default now()
)
`

	assert.Equal(t, expected, createTenantsTableSQL)
}

func TestBaseDialectGetCreateMigrationsTableSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := CreateDialect(config)

	createMigrationsTableSQL := dialect.GetCreateMigrationsTableSQL()

	expected := `
create table if not exists migrator.migrator_migrations (
  id serial primary key,
  name varchar(200) not null,
  source_dir varchar(200) not null,
  filename varchar(200) not null,
  type int not null,
  db_schema varchar(200) not null,
  created timestamp default now(),
	contents text,
	checksum varchar(64)
)
`

	assert.Equal(t, expected, createMigrationsTableSQL)
}

func TestBaseDialectGetCreateSchemaSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"

	dialect := CreateDialect(config)

	createSchemaSQL := dialect.GetCreateSchemaSQL("abc")

	expected := "create schema if not exists abc"

	assert.Equal(t, expected, createSchemaSQL)
}

func TestMSSQLDialectGetCreateSchemaSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"

	dialect := CreateDialect(config)

	createSchemaSQL := dialect.GetCreateSchemaSQL("def")

	expected := `
IF NOT EXISTS (select * from information_schema.schemata where schema_name = 'def')
BEGIN
  EXEC sp_executesql N'create schema def';
END
`

	assert.Equal(t, expected, createSchemaSQL)
}

func TestDoInitCannotBeginTransactionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)

	config := &config.Config{}
	config.Driver = "sqlmock"
	connector := BaseConnector{config, nil, nil}

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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	connector.DB = db

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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	// don't have to provide full SQL here - patterns at work
	mock.ExpectQuery("select").WillReturnError(errors.New("trouble maker"))

	connector.DB = db

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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	rows := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(rows)
	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	connector.DB = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	err = connector.ApplyMigrations(migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	err = connector.ApplyMigrations(migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	tenants := sqlmock.NewRows([]string{"name"}).AddRow("tenantname")
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	err = connector.ApplyMigrations(migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	tenant := "tenantname"
	tenants := sqlmock.NewRows([]string{"name"}).AddRow(tenant)
	mock.ExpectQuery("select").WillReturnRows(tenants)
	mock.ExpectBegin()
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	err = connector.ApplyMigrations(migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	mock.ExpectBegin().WillReturnError(errors.New("trouble maker tx.Begin()"))

	connector.DB = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	err = connector.AddTenantAndApplyMigrations("newtenant", migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	err = connector.AddTenantAndApplyMigrations("newtenant", migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	t1 := time.Now().UnixNano()
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{tenant1}

	err = connector.AddTenantAndApplyMigrations("newtenant", migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	tenant := "tenant"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	t1 := time.Now().UnixNano()
	m1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m1}

	err = connector.AddTenantAndApplyMigrations(tenant, migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	tenant := "tenant"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	t1 := time.Now().UnixNano()
	m1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m1}

	err = connector.AddTenantAndApplyMigrations(tenant, migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	tenant := "tenant"

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	t1 := time.Now().UnixNano()
	m1 := types.Migration{Name: fmt.Sprintf("%v.sql", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", t1), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m1}

	err = connector.AddTenantAndApplyMigrations(tenant, migrationsToApply)
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
	dialect := CreateDialect(config)
	connector := BaseConnector{config, dialect, nil}

	tenant := "tenant"
	time := time.Now().UnixNano()
	m := types.Migration{Name: fmt.Sprintf("%v.sql", time), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.sql", time), MigrationType: types.MigrationTypeTenantSchema, Contents: "insert into {schema}.settings values (456, '456') "}
	migrationsToApply := []types.Migration{m}

	mock.ExpectBegin()
	mock.ExpectExec("create schema").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(tenant).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into")
	mock.ExpectExec("insert into").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("insert into").ExpectExec().WithArgs(m.Name, m.SourceDir, m.File, m.MigrationType, tenant, m.Contents, m.CheckSum).WillReturnError(errors.New("trouble maker"))
	mock.ExpectRollback()

	connector.DB = db

	err = connector.AddTenantAndApplyMigrations(tenant, migrationsToApply)
	assert.Equal(t, "Failed to add migration entry: trouble maker", err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
