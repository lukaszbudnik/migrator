package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func getSupportedDatabases() []string {
	return []string{"postgresql", "mysql", "mariadb", "percona", "mssql"}
}

func TestGetTenants(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			tenants := connector.GetTenants()

			assert.True(t, len(tenants) >= 3)
			assert.Contains(t, tenants, types.Tenant{Name: "abc"})
			assert.Contains(t, tenants, types.Tenant{Name: "def"})
			assert.Contains(t, tenants, types.Tenant{Name: "xyz"})
		})
	}
}

func TestCreateVersion(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
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
		})
	}
}

func TestCreateVersionEmptyMigrationArray(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			migrationsToApply := []types.Migration{}

			results, version := connector.CreateVersion("commit-sha", types.ActionApply, false, migrationsToApply)
			// empty migrations slice - no version created
			assert.Nil(t, version)
			assert.Equal(t, int32(0), results.MigrationsGrandTotal)
			assert.Equal(t, int32(0), results.ScriptsGrandTotal)
		})
	}
}

func TestGetTenantsSQLDefault(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			dialect := newDialect(config)
			connector := baseConnector{newTestContext(), config, dialect, nil}
			defer connector.Dispose()

			tenantSelectSQL := connector.getTenantSelectSQL()

			assert.Equal(t, "select name from migrator.migrator_tenants", tenantSelectSQL)
		})
	}
}

func TestCreateTenant(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
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
		})
	}
}

func TestGetVersions(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			versions := connector.GetVersions()

			assert.True(t, len(versions) >= 2)
			// versions are sorted from newest (highest ID) to oldest (lowest ID)
			assert.True(t, versions[0].ID > versions[1].ID)
		})
	}
}

func TestGetVersionsByFile(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			versions := connector.GetVersions()
			existingVersion := versions[0]

			versions = connector.GetVersionsByFile(versions[0].DBMigrations[0].File)
			version := versions[0]
			assert.Equal(t, existingVersion.ID, version.ID)
			assert.Equal(t, existingVersion.DBMigrations[0].File, version.DBMigrations[0].File)
			assert.True(t, len(version.DBMigrations) > 0)
		})
	}
}

func TestGetVersionByID(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			versions := connector.GetVersions()
			existingVersion := versions[0]

			version, err := connector.GetVersionByID(existingVersion.ID)
			assert.Nil(t, err)
			assert.Equal(t, existingVersion.ID, version.ID)
			assert.True(t, len(version.DBMigrations) > 0)
		})
	}
}

func TestGetVersionByIDNotFound(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			version, err := connector.GetVersionByID(-1)
			assert.Nil(t, version)
			assert.Equal(t, "Version not found ID: -1", err.Error())
		})
	}
}

func TestGetDBMigrationByID(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			versions := connector.GetVersions()
			existingVersion := versions[0]
			existingDBMigration := existingVersion.DBMigrations[0]

			dbMigration, err := connector.GetDBMigrationByID(existingDBMigration.ID)
			assert.Nil(t, err)
			assert.Equal(t, existingDBMigration.ID, dbMigration.ID)
		})
	}
}

func TestGetDBMigrationByIDNotFound(t *testing.T) {
	supportedDatabases := getSupportedDatabases()

	for _, database := range supportedDatabases {
		t.Run(database, func(t *testing.T) {
			configFile := fmt.Sprintf("../test/migrator-%s.yaml", database)
			config, err := config.FromFile(configFile)
			assert.Nil(t, err)

			connector := New(newTestContext(), config)
			defer connector.Dispose()

			dbMigration, err := connector.GetDBMigrationByID(-1)
			assert.Nil(t, dbMigration)
			assert.Equal(t, "DB migration not found ID: -1", err.Error())
		})
	}
}
