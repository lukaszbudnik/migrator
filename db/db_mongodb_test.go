package db

import (
	"context"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoDBConnectorCreation(t *testing.T) {
	config := &config.Config{
		Driver:     "mongodb",
		DataSource: "mongodb://localhost:27017",
	}

	connector := newMongoDBConnector(context.Background(), config)
	assert.NotNil(t, connector)

	mongoConnector, ok := connector.(*mongoDBConnector)
	assert.True(t, ok)
	assert.Equal(t, "mongodb://localhost:27017", mongoConnector.config.DataSource)
	assert.False(t, mongoConnector.initialised)
}

func TestMongoDBConnectorFactory(t *testing.T) {
	config := &config.Config{
		Driver:     "mongodb",
		DataSource: "mongodb://localhost:27017",
	}

	connector := New(context.Background(), config)
	assert.NotNil(t, connector)

	_, ok := connector.(*mongoDBConnector)
	assert.True(t, ok, "Expected mongoDBConnector type")
}

func TestMongoDBConnectorFactoryFallback(t *testing.T) {
	config := &config.Config{
		Driver:     "postgres",
		DataSource: "user=test",
	}

	connector := New(context.Background(), config)
	assert.NotNil(t, connector)

	_, ok := connector.(*baseConnector)
	assert.True(t, ok, "Expected baseConnector type for non-MongoDB driver")
}

func TestMongoDBCollectionNames(t *testing.T) {
	assert.Equal(t, "migrator", migratorSchema)
	assert.Equal(t, "migrator_tenants", migratorTenantsTable)
	assert.Equal(t, "migrator_migrations", migratorMigrationsTable)
	assert.Equal(t, "migrator_versions", migratorVersionsTable)
}

func TestMongoDBDocToDBMigration(t *testing.T) {
	config := &config.Config{
		Driver:     "mongodb",
		DataSource: "mongodb://localhost:27017",
	}

	connector := newMongoDBConnector(context.Background(), config)
	mongoConnector := connector.(*mongoDBConnector)

	now := time.Now()
	doc := bson.M{
		"_id":        int32(123),
		"name":       "001_test.js",
		"source_dir": "tenants",
		"filename":   "tenants/001_test.js",
		"type":       int32(2),
		"db_schema":  "test_tenant",
		"created":    now,
		"contents":   "db.users.insertOne({name: 'test'})",
		"checksum":   "abc123",
	}

	migration := mongoConnector.docToDBMigration(doc)

	assert.Equal(t, int32(123), migration.ID)
	assert.Equal(t, "001_test.js", migration.Name)
	assert.Equal(t, "tenants", migration.SourceDir)
	assert.Equal(t, "tenants/001_test.js", migration.File)
	assert.Equal(t, types.MigrationTypeTenantMigration, migration.MigrationType)
	assert.Equal(t, "test_tenant", migration.Schema)
	assert.Equal(t, "db.users.insertOne({name: 'test'})", migration.Contents)
	assert.Equal(t, "abc123", migration.CheckSum)
	assert.Equal(t, now, migration.Created.Time)
}

func TestMongoDBComputeSummary(t *testing.T) {
	config := &config.Config{
		Driver:     "mongodb",
		DataSource: "mongodb://localhost:27017",
	}

	connector := newMongoDBConnector(context.Background(), config)
	mongoConnector := connector.(*mongoDBConnector)

	summary := &types.Summary{
		StartedAt: graphql.Time{Time: time.Now()},
		Tenants:   3,
	}

	migrations := []types.Migration{
		{MigrationType: types.MigrationTypeSingleMigration},
		{MigrationType: types.MigrationTypeSingleMigration},
		{MigrationType: types.MigrationTypeTenantMigration},
		{MigrationType: types.MigrationTypeTenantMigration},
		{MigrationType: types.MigrationTypeTenantMigration},
		{MigrationType: types.MigrationTypeSingleScript},
		{MigrationType: types.MigrationTypeTenantScript},
	}

	tenants := []types.Tenant{
		{Name: "tenant1"},
		{Name: "tenant2"},
		{Name: "tenant3"},
	}

	mongoConnector.computeSummary(summary, migrations, tenants)

	assert.Equal(t, int32(2), summary.SingleMigrations)
	assert.Equal(t, int32(3), summary.TenantMigrations)
	assert.Equal(t, int32(1), summary.SingleScripts)
	assert.Equal(t, int32(1), summary.TenantScripts)
	assert.Equal(t, int32(9), summary.TenantMigrationsTotal) // 3 tenants * 3 migrations
	assert.Equal(t, int32(3), summary.TenantScriptsTotal)    // 3 tenants * 1 script
	assert.Equal(t, int32(11), summary.MigrationsGrandTotal) // 2 + 9
	assert.Equal(t, int32(4), summary.ScriptsGrandTotal)     // 1 + 3
}

func TestMongoDBSchemaPlaceholderDefault(t *testing.T) {
	config := &config.Config{
		Driver:     "mongodb",
		DataSource: "mongodb://localhost:27017",
	}

	connector := newMongoDBConnector(context.Background(), config)
	mongoConnector := connector.(*mongoDBConnector)

	migration := types.Migration{
		Contents: "db.getSiblingDB('{schema}').users.insertOne({name: 'test'})",
	}

	// Simulate what executeMigration does with placeholder replacement
	schemaPlaceHolder := mongoConnector.config.SchemaPlaceHolder
	if schemaPlaceHolder == "" {
		schemaPlaceHolder = defaultSchemaPlaceHolder
	}

	assert.Equal(t, "{schema}", schemaPlaceHolder)
	assert.Contains(t, migration.Contents, "{schema}")
}

func TestMongoDBSchemaPlaceholderCustom(t *testing.T) {
	config := &config.Config{
		Driver:            "mongodb",
		DataSource:        "mongodb://localhost:27017",
		SchemaPlaceHolder: ":tenant",
	}

	connector := newMongoDBConnector(context.Background(), config)
	mongoConnector := connector.(*mongoDBConnector)

	assert.Equal(t, ":tenant", mongoConnector.config.SchemaPlaceHolder)

	migration := types.Migration{
		Contents: "db.getSiblingDB(':tenant').users.insertOne({name: 'test'})",
	}

	assert.Contains(t, migration.Contents, ":tenant")
}

func TestMongoDBDisposeNilClient(t *testing.T) {
	config := &config.Config{
		Driver:     "mongodb",
		DataSource: "mongodb://localhost:27017",
	}

	connector := newMongoDBConnector(context.Background(), config)
	mongoConnector := connector.(*mongoDBConnector)

	// Client is nil before init
	assert.Nil(t, mongoConnector.client)

	// Should not panic
	connector.Dispose()
}

func TestMongoDBTenantDocumentStructure(t *testing.T) {
	// Test that we can create a tenant document structure
	tenantDoc := bson.M{
		"name":    "test_tenant",
		"created": time.Now(),
	}

	assert.Equal(t, "test_tenant", tenantDoc["name"])
	assert.NotNil(t, tenantDoc["created"])
}

func TestMongoDBVersionDocumentStructure(t *testing.T) {
	// Test that we can create a version document structure
	versionDoc := bson.M{
		"_id":     int32(1),
		"name":    "v1.0.0",
		"created": time.Now(),
	}

	assert.Equal(t, int32(1), versionDoc["_id"])
	assert.Equal(t, "v1.0.0", versionDoc["name"])
	assert.NotNil(t, versionDoc["created"])
}

func TestMongoDBMigrationDocumentStructure(t *testing.T) {
	// Test that we can create a migration document structure
	migrationDoc := bson.M{
		"_id":        int32(1),
		"name":       "001_init.js",
		"source_dir": "admin",
		"filename":   "admin/001_init.js",
		"type":       int32(1),
		"db_schema":  "migrator",
		"created":    time.Now(),
		"contents":   "db.config.insertOne({key: 'value'})",
		"checksum":   "abc123",
		"version_id": int32(1),
	}

	assert.Equal(t, int32(1), migrationDoc["_id"])
	assert.Equal(t, "001_init.js", migrationDoc["name"])
	assert.Equal(t, "admin", migrationDoc["source_dir"])
	assert.Equal(t, "admin/001_init.js", migrationDoc["filename"])
	assert.Equal(t, int32(1), migrationDoc["type"])
	assert.Equal(t, "migrator", migrationDoc["db_schema"])
	assert.Equal(t, "db.config.insertOne({key: 'value'})", migrationDoc["contents"])
	assert.Equal(t, "abc123", migrationDoc["checksum"])
	assert.Equal(t, int32(1), migrationDoc["version_id"])
}

func TestMongoDBCounterDocumentStructure(t *testing.T) {
	// Test that we can create a counter document structure
	counterDoc := bson.M{
		"_id": "version_id",
		"seq": int32(5),
	}

	assert.Equal(t, "version_id", counterDoc["_id"])
	assert.Equal(t, int32(5), counterDoc["seq"])
}

func TestMongoDBInitIdempotent(t *testing.T) {
	config := &config.Config{
		Driver:     "mongodb",
		DataSource: "mongodb://localhost:27017",
	}

	connector := newMongoDBConnector(context.Background(), config)
	mongoConnector := connector.(*mongoDBConnector)

	// Before init
	assert.False(t, mongoConnector.initialised)

	// Calling init multiple times should be safe (though will fail without DB)
	// We just test the idempotent flag logic
	mongoConnector.initialised = true

	err := mongoConnector.init()
	assert.Nil(t, err) // Should return nil immediately when already initialized
	assert.True(t, mongoConnector.initialised)
}
