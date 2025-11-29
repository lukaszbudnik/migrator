package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func TestMongoDBGetTenants(t *testing.T) {
	configFile := "../test/migrator-mongodb.yaml"
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	// Create test tenants
	connector.CreateTenant("abc", "test-tenant-abc", types.ActionSync, []types.Migration{}, false)
	connector.CreateTenant("def", "test-tenant-def", types.ActionSync, []types.Migration{}, false)
	connector.CreateTenant("xyz", "test-tenant-xyz", types.ActionSync, []types.Migration{}, false)

	tenants := connector.GetTenants()

	assert.True(t, len(tenants) >= 3)
	assert.Contains(t, tenants, types.Tenant{Name: "abc"})
	assert.Contains(t, tenants, types.Tenant{Name: "def"})
	assert.Contains(t, tenants, types.Tenant{Name: "xyz"})
}

func TestMongoDBCreateVersion(t *testing.T) {
	configFile := "../test/migrator-mongodb.yaml"
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	// Create test tenants
	connector.CreateTenant("tenant1", "test-tenant-1", types.ActionSync, []types.Migration{}, false)
	connector.CreateTenant("tenant2", "test-tenant-2", types.ActionSync, []types.Migration{}, false)

	tenants := connector.GetTenants()
	noOfTenants := len(tenants)

	dbMigrationsBefore := connector.GetAppliedMigrations()
	lenBefore := len(dbMigrationsBefore)

	p1 := time.Now().UnixNano()
	p2 := time.Now().UnixNano()
	c1 := time.Now().UnixNano()
	t1 := time.Now().UnixNano()
	t2 := time.Now().UnixNano()

	// ref migrations - use unique filenames to avoid duplicates
	ref1 := types.Migration{Name: fmt.Sprintf("%v.js", p1), SourceDir: "ref", File: fmt.Sprintf("ref/%v.js", p1), MigrationType: types.MigrationTypeSingleMigration, Contents: "db.modules.insertOne({k: 123, v: '123'})"}
	ref2 := types.Migration{Name: fmt.Sprintf("%v.js", p2), SourceDir: "ref", File: fmt.Sprintf("ref/%v.js", p2), MigrationType: types.MigrationTypeSingleMigration, Contents: "db.modules.insertOne({k: 456, v: '456'})"}

	// configuration migrations - use unique filenames to avoid duplicates
	config1 := types.Migration{Name: fmt.Sprintf("%v.js", c1), SourceDir: "configuration", File: fmt.Sprintf("configuration/%v.js", c1), MigrationType: types.MigrationTypeSingleMigration, Contents: "db.feature_toggles.insertOne({name: 'dark_mode_" + fmt.Sprintf("%v", c1) + "', enabled: true})"}

	// tenant migrations - use unique filenames to avoid duplicates
	tenant1 := types.Migration{Name: fmt.Sprintf("%v.js", t1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.js", t1), MigrationType: types.MigrationTypeTenantMigration, Contents: "db.settings.insertOne({k: 789, v: '789'})"}
	tenant2 := types.Migration{Name: fmt.Sprintf("%v.js", t2), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.js", t2), MigrationType: types.MigrationTypeTenantMigration, Contents: "db.settings.insertOne({k: 101, v: '101'})"}

	migrationsToApply := []types.Migration{ref1, ref2, config1, tenant1, tenant2}

	results, version := connector.CreateVersion("commit-sha-mongo", types.ActionApply, migrationsToApply, false)

	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, "commit-sha-mongo", version.Name)
	assert.Equal(t, results.MigrationsGrandTotal, int32(len(version.DBMigrations)))
	assert.Equal(t, int32(noOfTenants), results.Tenants)
	assert.Equal(t, int32(3), results.SingleMigrations)
	assert.Equal(t, int32(2), results.TenantMigrations)
	assert.Equal(t, int32(noOfTenants*2), results.TenantMigrationsTotal)
	assert.Equal(t, int32(noOfTenants*2+3), results.MigrationsGrandTotal)

	dbMigrationsAfter := connector.GetAppliedMigrations()
	lenAfter := len(dbMigrationsAfter)

	assert.Equal(t, lenBefore+int(results.MigrationsGrandTotal), lenAfter)
}

func TestMongoDBScripts(t *testing.T) {
	configFile := "../test/migrator-mongodb.yaml"
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	// Create test tenant with some initial data
	testTenant := fmt.Sprintf("scripttenant%d", time.Now().UnixNano())
	m1 := time.Now().UnixNano()
	tenantMigration := types.Migration{Name: fmt.Sprintf("%v.js", m1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.js", m1), MigrationType: types.MigrationTypeTenantMigration, Contents: "db.settings.insertOne({k: 999, v: '999'})"}
	connector.CreateTenant(testTenant, "test-tenant-scripts", types.ActionApply, []types.Migration{tenantMigration}, false)

	tenants := connector.GetTenants()
	noOfTenants := len(tenants)

	s1 := time.Now().UnixNano()
	ts1 := time.Now().UnixNano()

	// Single script - runs every time, uses getSiblingDB to access ref database
	singleScript := types.Migration{Name: fmt.Sprintf("%v.js", s1), SourceDir: "ref-scripts", File: fmt.Sprintf("ref-scripts/%v.js", s1), MigrationType: types.MigrationTypeSingleScript, Contents: "db.getSiblingDB('ref').modules.updateMany({}, {$set: {script_updated: new Date()}})"}

	// Tenant script - runs every time for each tenant, uses updateMany
	tenantScript := types.Migration{Name: fmt.Sprintf("%v.js", ts1), SourceDir: "tenants-scripts", File: fmt.Sprintf("tenants-scripts/%v.js", ts1), MigrationType: types.MigrationTypeTenantScript, Contents: "db.settings.updateMany({}, {$set: {script_run_at: new Date()}})"}

	scriptsToApply := []types.Migration{singleScript, tenantScript}

	results, version := connector.CreateVersion("test-scripts", types.ActionApply, scriptsToApply, false)

	assert.NotNil(t, version)
	assert.Equal(t, int32(1), results.SingleScripts)
	assert.Equal(t, int32(1), results.TenantScripts)
	assert.Equal(t, int32(noOfTenants), results.TenantScriptsTotal)
	assert.Equal(t, int32(1+noOfTenants), results.ScriptsGrandTotal)
}

func TestMongoDBCreateTenant(t *testing.T) {
	configFile := "../test/migrator-mongodb.yaml"
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	tenantsBefore := connector.GetTenants()
	lenBefore := len(tenantsBefore)

	p1 := time.Now().UnixNano()
	p2 := time.Now().UnixNano()

	tenant1 := types.Migration{Name: fmt.Sprintf("%v.js", p1), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.js", p1), MigrationType: types.MigrationTypeTenantMigration, Contents: "db.data.insertOne({k: 111, v: '111'})"}
	tenant2 := types.Migration{Name: fmt.Sprintf("%v.js", p2), SourceDir: "tenants", File: fmt.Sprintf("tenants/%v.js", p2), MigrationType: types.MigrationTypeTenantMigration, Contents: "db.data.insertOne({k: 222, v: '222'})"}

	migrationsToApply := []types.Migration{tenant1, tenant2}

	newTenantName := fmt.Sprintf("newtenant%d", time.Now().UnixNano())
	results, version := connector.CreateTenant(newTenantName, "create-tenant-version", types.ActionApply, migrationsToApply, false)

	assert.NotNil(t, version)
	assert.True(t, version.ID > 0)
	assert.Equal(t, int32(1), results.Tenants)
	assert.Equal(t, int32(2), results.TenantMigrations)
	assert.Equal(t, int32(2), results.TenantMigrationsTotal)
	assert.Equal(t, int32(2), results.MigrationsGrandTotal)

	tenantsAfter := connector.GetTenants()
	lenAfter := len(tenantsAfter)

	assert.Equal(t, lenBefore+1, lenAfter)
	assert.Contains(t, tenantsAfter, types.Tenant{Name: newTenantName})
}

func TestMongoDBHealthCheck(t *testing.T) {
	configFile := "../test/migrator-mongodb.yaml"
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	connector := New(newTestContext(), config)
	defer connector.Dispose()

	err = connector.HealthCheck()
	assert.Nil(t, err)
}

func TestMongoDBCustomTenantCollection(t *testing.T) {
	// Test with custom collection name (default field)
	cfg := &config.Config{
		Driver:           "mongodb",
		DataSource:       "mongodb://localhost:27017",
		TenantSelect:     "custom_tenants",
		TenantInsert:     "custom_tenants",
		SingleMigrations: []string{"admin"},
	}

	connector := New(newTestContext(), cfg)
	defer connector.Dispose()

	// Create tenant in custom collection
	tenantName := fmt.Sprintf("custom_tenant_%d", time.Now().UnixNano())
	results, version := connector.CreateTenant(tenantName, "test-custom-collection", types.ActionSync, []types.Migration{}, false)

	assert.NotNil(t, version)
	assert.Equal(t, int32(1), results.Tenants)

	// Verify tenant was created in custom collection
	tenants := connector.GetTenants()
	assert.Contains(t, tenants, types.Tenant{Name: tenantName})
}

func TestMongoDBCustomTenantCollectionAndField(t *testing.T) {
	// Test with custom collection and custom field name
	cfg := &config.Config{
		Driver:           "mongodb",
		DataSource:       "mongodb://localhost:27017",
		TenantSelect:     "organizations.org_name",
		TenantInsert:     "organizations.org_name",
		SingleMigrations: []string{"admin"},
	}

	connector := New(newTestContext(), cfg)
	defer connector.Dispose()

	// Create tenant in custom collection with custom field
	tenantName := fmt.Sprintf("org_%d", time.Now().UnixNano())
	results, version := connector.CreateTenant(tenantName, "test-custom-field", types.ActionSync, []types.Migration{}, false)

	assert.NotNil(t, version)
	assert.Equal(t, int32(1), results.Tenants)

	// Verify tenant was created with custom field
	tenants := connector.GetTenants()
	assert.Contains(t, tenants, types.Tenant{Name: tenantName})
}

func TestMongoDBBackwardCompatibilityTenantSelectSQL(t *testing.T) {
	// Test that old tenantSelectSQL field still works
	cfg := &config.Config{
		Driver:           "mongodb",
		DataSource:       "mongodb://localhost:27017",
		TenantSelectSQL:  "legacy_tenants",
		TenantInsertSQL:  "legacy_tenants",
		SingleMigrations: []string{"admin"},
	}

	connector := New(newTestContext(), cfg)
	defer connector.Dispose()

	// Create tenant using old config field names
	tenantName := fmt.Sprintf("legacy_tenant_%d", time.Now().UnixNano())
	results, version := connector.CreateTenant(tenantName, "test-legacy-config", types.ActionSync, []types.Migration{}, false)

	assert.NotNil(t, version)
	assert.Equal(t, int32(1), results.Tenants)

	// Verify tenant was created
	tenants := connector.GetTenants()
	assert.Contains(t, tenants, types.Tenant{Name: tenantName})
}
