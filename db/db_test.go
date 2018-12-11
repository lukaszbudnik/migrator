package db

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateConnectorPanicUnknownDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "abcxyz"

	assert.Panics(t, func() {
		CreateConnector(config)
	}, "Should panic because of unknown driver")
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
	assert.Panics(t, func() {
		connector.Init()
	}, "Should panic because of connection error")
}

func TestDBGetTenantsPanicSQLSyntaxError(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.TenantSelectSQL = "sadfdsfdsf"
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
	m := types.Migration{Name: "201602220002.sql", SourceDir: "error", File: "error/201602220002.sql", MigrationType: types.MigrationTypeTenantSchema, Contents: "createtablexyx ( idint primary key (id) )"}
	ms := []types.Migration{m}

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

	dbMigrationsBefore := connector.GetDBMigrations()
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

	dbMigrationsAfter := connector.GetDBMigrations()
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

	dbMigrationsBefore := connector.GetDBMigrations()
	lenBefore := len(dbMigrationsBefore)

	migrationsToApply := []types.Migration{}

	connector.ApplyMigrations(migrationsToApply)

	dbMigrationsAfter := connector.GetDBMigrations()
	lenAfter := len(dbMigrationsAfter)

	assert.Equal(t, lenAfter, lenBefore)
}

func TestGetTenantsSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantSelectSQL := connector.GetTenantSelectSQL()

	assert.Equal(t, "select name from migrator.migrator_tenants", tenantSelectSQL)
}

func TestGetTenantsSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantSelectSQL := connector.GetTenantSelectSQL()

	assert.Equal(t, "select somename from someschema.sometable", tenantSelectSQL)
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

	dbMigrationsBefore := connector.GetDBMigrations()
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

	dbMigrationsAfter := connector.GetDBMigrations()
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
	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantInsertSQL := connector.GetTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values (?)", tenantInsertSQL)
}

func TestPostgreSQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "postgres"
	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantInsertSQL := connector.GetTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values ($1)", tenantInsertSQL)
}

func TestMSSQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "sqlserver"
	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantInsertSQL := connector.GetTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values (@p1)", tenantInsertSQL)
}

func TestGetTenantInsertSQLOverride(t *testing.T) {
	config, err := config.FromFile("../test/migrator-overrides.yaml")
	assert.Nil(t, err)

	connector := CreateConnector(config)
	defer connector.Dispose()

	tenantInsertSQL := connector.GetTenantInsertSQL()

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
