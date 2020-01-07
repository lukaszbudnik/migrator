package db

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestDBCreateDialectMysqlDriver(t *testing.T) {
	config := &config.Config{}
	config.Driver = "mysql"
	dialect := newDialect(config)
	assert.IsType(t, &mySQLDialect{}, dialect)
}

func TestMySQLGetMigrationInsertSQL(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"

	dialect := newDialect(config)

	insertMigrationSQL := dialect.GetMigrationInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_migrations (name, source_dir, filename, type, db_schema, contents, checksum) values (?, ?, ?, ?, ?, ?, ?)", insertMigrationSQL)
}

func TestMySQLGetTenantInsertSQLDefault(t *testing.T) {
	config, err := config.FromFile("../test/migrator.yaml")
	assert.Nil(t, err)

	config.Driver = "mysql"
	dialect := newDialect(config)
	connector := baseConnector{newTestContext(), config, dialect, nil}
	defer connector.Dispose()

	tenantInsertSQL := connector.getTenantInsertSQL()

	assert.Equal(t, "insert into migrator.migrator_tenants (name) values (?)", tenantInsertSQL)
}
