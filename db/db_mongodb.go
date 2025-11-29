package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDBConnector struct {
	ctx         context.Context
	config      *config.Config
	client      *mongo.Client
	db          *mongo.Database
	initialised bool
}

func newMongoDBConnector(ctx context.Context, config *config.Config) Connector {
	return &mongoDBConnector{
		ctx:         ctx,
		config:      config,
		initialised: false,
	}
}

func (mc *mongoDBConnector) init() error {
	if mc.initialised {
		return nil
	}

	clientOptions := options.Client().ApplyURI(mc.config.DataSource)
	client, err := mongo.Connect(mc.ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(mc.ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	mc.client = client
	mc.db = client.Database(migratorSchema)

	// Create collections and indexes
	if err := mc.createCollections(); err != nil {
		return err
	}

	mc.initialised = true
	return nil
}

func (mc *mongoDBConnector) createCollections() error {
	// Create tenants collection
	tenantsCol := mc.db.Collection(migratorTenantsTable)
	_, err := tenantsCol.Indexes().CreateOne(mc.ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create tenants index: %v", err)
	}

	// Create versions collection
	versionsCol := mc.db.Collection(migratorVersionsTable)
	_, err = versionsCol.Indexes().CreateOne(mc.ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "created", Value: -1}},
	})
	if err != nil {
		return fmt.Errorf("failed to create versions index: %v", err)
	}

	// Create migrations collection
	migrationsCol := mc.db.Collection(migratorMigrationsTable)
	_, err = migrationsCol.Indexes().CreateOne(mc.ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "version_id", Value: 1}},
	})
	if err != nil {
		return fmt.Errorf("failed to create migrations index: %v", err)
	}

	return nil
}

func (mc *mongoDBConnector) GetTenants() []types.Tenant {
	if err := mc.init(); err != nil {
		common.LogError(mc.ctx, "Failed to initialize MongoDB: %v", err)
		return []types.Tenant{}
	}

	col := mc.db.Collection(migratorTenantsTable)
	cursor, err := col.Find(mc.ctx, bson.M{})
	if err != nil {
		common.LogError(mc.ctx, "Failed to get tenants: %v", err)
		return []types.Tenant{}
	}
	defer cursor.Close(mc.ctx)

	var tenants []types.Tenant
	for cursor.Next(mc.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		tenants = append(tenants, types.Tenant{Name: doc["name"].(string)})
	}

	return tenants
}

func (mc *mongoDBConnector) GetVersions() []types.Version {
	if err := mc.init(); err != nil {
		common.LogError(mc.ctx, "Failed to initialize MongoDB: %v", err)
		return []types.Version{}
	}

	versionsCol := mc.db.Collection(migratorVersionsTable)
	migrationsCol := mc.db.Collection(migratorMigrationsTable)

	cursor, err := versionsCol.Find(mc.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}))
	if err != nil {
		common.LogError(mc.ctx, "Failed to get versions: %v", err)
		return []types.Version{}
	}
	defer cursor.Close(mc.ctx)

	var versions []types.Version
	for cursor.Next(mc.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		versionID := doc["_id"].(int32)
		version := types.Version{
			ID:      versionID,
			Name:    doc["name"].(string),
			Created: graphql.Time{Time: mc.convertToTime(doc["created"])},
		}

		// Get migrations for this version
		migCursor, err := migrationsCol.Find(mc.ctx, bson.M{"version_id": versionID})
		if err == nil {
			for migCursor.Next(mc.ctx) {
				var migDoc bson.M
				if err := migCursor.Decode(&migDoc); err != nil {
					continue
				}
				version.DBMigrations = append(version.DBMigrations, mc.docToDBMigration(migDoc))
			}
			migCursor.Close(mc.ctx)
		}

		versions = append(versions, version)
	}

	return versions
}

func (mc *mongoDBConnector) GetVersionsByFile(file string) []types.Version {
	if err := mc.init(); err != nil {
		common.LogError(mc.ctx, "Failed to initialize MongoDB: %v", err)
		return []types.Version{}
	}

	migrationsCol := mc.db.Collection(migratorMigrationsTable)
	cursor, err := migrationsCol.Find(mc.ctx, bson.M{"filename": file})
	if err != nil {
		return []types.Version{}
	}
	defer cursor.Close(mc.ctx)

	versionIDs := make(map[int32]bool)
	for cursor.Next(mc.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		versionIDs[doc["version_id"].(int32)] = true
	}

	var versions []types.Version
	for versionID := range versionIDs {
		if version, err := mc.GetVersionByID(versionID); err == nil {
			versions = append(versions, *version)
		}
	}

	return versions
}

func (mc *mongoDBConnector) GetVersionByID(ID int32) (*types.Version, error) {
	if err := mc.init(); err != nil {
		return nil, err
	}

	versionsCol := mc.db.Collection(migratorVersionsTable)
	var doc bson.M
	err := versionsCol.FindOne(mc.ctx, bson.M{"_id": ID}).Decode(&doc)
	if err != nil {
		return nil, err
	}

	version := &types.Version{
		ID:      doc["_id"].(int32),
		Name:    doc["name"].(string),
		Created: graphql.Time{Time: mc.convertToTime(doc["created"])},
	}

	migrationsCol := mc.db.Collection(migratorMigrationsTable)
	cursor, err := migrationsCol.Find(mc.ctx, bson.M{"version_id": ID})
	if err == nil {
		defer cursor.Close(mc.ctx)
		for cursor.Next(mc.ctx) {
			var migDoc bson.M
			if err := cursor.Decode(&migDoc); err != nil {
				continue
			}
			version.DBMigrations = append(version.DBMigrations, mc.docToDBMigration(migDoc))
		}
	}

	return version, nil
}

func (mc *mongoDBConnector) GetDBMigrationByID(ID int32) (*types.DBMigration, error) {
	if err := mc.init(); err != nil {
		return nil, err
	}

	col := mc.db.Collection(migratorMigrationsTable)
	var doc bson.M
	err := col.FindOne(mc.ctx, bson.M{"_id": ID}).Decode(&doc)
	if err != nil {
		return nil, err
	}

	migration := mc.docToDBMigration(doc)
	return &migration, nil
}

func (mc *mongoDBConnector) GetAppliedMigrations() []types.DBMigration {
	if err := mc.init(); err != nil {
		common.LogError(mc.ctx, "Failed to initialize MongoDB: %v", err)
		return []types.DBMigration{}
	}

	col := mc.db.Collection(migratorMigrationsTable)
	cursor, err := col.Find(mc.ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}, {Key: "source_dir", Value: 1}}))
	if err != nil {
		common.LogError(mc.ctx, "Failed to get applied migrations: %v", err)
		return []types.DBMigration{}
	}
	defer cursor.Close(mc.ctx)

	var migrations []types.DBMigration
	for cursor.Next(mc.ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		migrations = append(migrations, mc.docToDBMigration(doc))
	}

	return migrations
}

func (mc *mongoDBConnector) CreateVersion(versionName string, action types.Action, migrations []types.Migration, dryRun bool) (*types.Summary, *types.Version) {
	if err := mc.init(); err != nil {
		common.LogError(mc.ctx, "Failed to initialize MongoDB: %v", err)
		return nil, nil
	}

	startTime := time.Now()
	tenants := mc.GetTenants()

	summary := &types.Summary{
		StartedAt: graphql.Time{Time: startTime},
		Tenants:   int32(len(tenants)),
	}

	if dryRun {
		mc.computeSummary(summary, migrations, tenants)
		summary.Duration = time.Since(startTime).Seconds()
		return summary, nil
	}

	// Create version
	versionsCol := mc.db.Collection(migratorVersionsTable)
	versionID := mc.getNextSequence("version_id")
	versionDoc := bson.M{
		"_id":     versionID,
		"name":    versionName,
		"created": time.Now(),
	}
	_, err := versionsCol.InsertOne(mc.ctx, versionDoc)
	if err != nil {
		common.LogError(mc.ctx, "Failed to create version: %v", err)
		return nil, nil
	}

	version := &types.Version{
		ID:      versionID,
		Name:    versionName,
		Created: graphql.Time{Time: time.Now()},
	}

	// Apply migrations
	for _, migration := range migrations {
		if migration.MigrationType == types.MigrationTypeSingleMigration || migration.MigrationType == types.MigrationTypeSingleScript {
			if action == types.ActionApply {
				mc.executeMigration(migration, migratorSchema)
			}
			mc.recordMigration(versionID, migration, migratorSchema, version)
			if migration.MigrationType == types.MigrationTypeSingleMigration {
				summary.SingleMigrations++
			} else {
				summary.SingleScripts++
			}
		} else {
			for _, tenant := range tenants {
				if action == types.ActionApply {
					mc.executeMigration(migration, tenant.Name)
				}
				mc.recordMigration(versionID, migration, tenant.Name, version)
			}
			if migration.MigrationType == types.MigrationTypeTenantMigration {
				summary.TenantMigrations++
			} else {
				summary.TenantScripts++
			}
		}
	}

	summary.TenantMigrationsTotal = summary.TenantMigrations * summary.Tenants
	summary.TenantScriptsTotal = summary.TenantScripts * summary.Tenants
	summary.MigrationsGrandTotal = summary.SingleMigrations + summary.TenantMigrationsTotal
	summary.ScriptsGrandTotal = summary.SingleScripts + summary.TenantScriptsTotal
	summary.Duration = time.Since(startTime).Seconds()
	summary.VersionID = versionID

	return summary, version
}

func (mc *mongoDBConnector) CreateTenant(tenantName string, versionName string, action types.Action, migrations []types.Migration, dryRun bool) (*types.Summary, *types.Version) {
	if err := mc.init(); err != nil {
		common.LogError(mc.ctx, "Failed to initialize MongoDB: %v", err)
		return nil, nil
	}

	startTime := time.Now()

	summary := &types.Summary{
		StartedAt: graphql.Time{Time: startTime},
		Tenants:   1,
	}

	if dryRun {
		for _, migration := range migrations {
			if migration.MigrationType == types.MigrationTypeTenantMigration {
				summary.TenantMigrations++
			} else if migration.MigrationType == types.MigrationTypeTenantScript {
				summary.TenantScripts++
			}
		}
		summary.TenantMigrationsTotal = summary.TenantMigrations
		summary.TenantScriptsTotal = summary.TenantScripts
		summary.MigrationsGrandTotal = summary.TenantMigrationsTotal
		summary.ScriptsGrandTotal = summary.TenantScriptsTotal
		summary.Duration = time.Since(startTime).Seconds()
		return summary, nil
	}

	// Create tenant
	tenantsCol := mc.db.Collection(migratorTenantsTable)
	_, err := tenantsCol.InsertOne(mc.ctx, bson.M{"name": tenantName, "created": time.Now()})
	if err != nil {
		common.LogError(mc.ctx, "Failed to create tenant: %v", err)
		return nil, nil
	}

	// Create version
	versionsCol := mc.db.Collection(migratorVersionsTable)
	versionID := mc.getNextSequence("version_id")
	versionDoc := bson.M{
		"_id":     versionID,
		"name":    versionName,
		"created": time.Now(),
	}
	_, err = versionsCol.InsertOne(mc.ctx, versionDoc)
	if err != nil {
		common.LogError(mc.ctx, "Failed to create version: %v", err)
		return nil, nil
	}

	version := &types.Version{
		ID:      versionID,
		Name:    versionName,
		Created: graphql.Time{Time: time.Now()},
	}

	// Apply tenant migrations
	for _, migration := range migrations {
		if migration.MigrationType == types.MigrationTypeTenantMigration || migration.MigrationType == types.MigrationTypeTenantScript {
			if action == types.ActionApply {
				mc.executeMigration(migration, tenantName)
			}
			mc.recordMigration(versionID, migration, tenantName, version)
			if migration.MigrationType == types.MigrationTypeTenantMigration {
				summary.TenantMigrations++
			} else {
				summary.TenantScripts++
			}
		}
	}

	summary.TenantMigrationsTotal = summary.TenantMigrations
	summary.TenantScriptsTotal = summary.TenantScripts
	summary.MigrationsGrandTotal = summary.TenantMigrationsTotal
	summary.ScriptsGrandTotal = summary.TenantScriptsTotal
	summary.Duration = time.Since(startTime).Seconds()
	summary.VersionID = versionID

	return summary, version
}

func (mc *mongoDBConnector) HealthCheck() error {
	if mc.client == nil {
		return mc.init()
	}
	return mc.client.Ping(mc.ctx, nil)
}

func (mc *mongoDBConnector) Dispose() {
	if mc.client != nil {
		mc.client.Disconnect(mc.ctx)
	}
}

// Helper methods

func (mc *mongoDBConnector) convertToTime(v interface{}) time.Time {
	switch t := v.(type) {
	case time.Time:
		return t
	case primitive.DateTime:
		return t.Time()
	default:
		return time.Time{}
	}
}

func (mc *mongoDBConnector) executeMigration(migration types.Migration, dbName string) {
	targetDB := mc.client.Database(dbName)

	// Replace schema placeholder
	schemaPlaceHolder := mc.config.SchemaPlaceHolder
	if schemaPlaceHolder == "" {
		schemaPlaceHolder = defaultSchemaPlaceHolder
	}
	contents := strings.ReplaceAll(migration.Contents, schemaPlaceHolder, dbName)

	// Execute as JavaScript
	var result bson.M
	err := targetDB.RunCommand(mc.ctx, bson.D{{Key: "eval", Value: contents}}).Decode(&result)
	if err != nil {
		common.LogError(mc.ctx, "Failed to execute migration %s: %v", migration.File, err)
	}
}

func (mc *mongoDBConnector) recordMigration(versionID int32, migration types.Migration, schema string, version *types.Version) {
	col := mc.db.Collection(migratorMigrationsTable)
	migrationID := mc.getNextSequence("migration_id")

	doc := bson.M{
		"_id":        migrationID,
		"name":       migration.Name,
		"source_dir": migration.SourceDir,
		"filename":   migration.File,
		"type":       int(migration.MigrationType),
		"db_schema":  schema,
		"created":    time.Now(),
		"contents":   migration.Contents,
		"checksum":   migration.CheckSum,
		"version_id": versionID,
	}

	_, err := col.InsertOne(mc.ctx, doc)
	if err != nil {
		common.LogError(mc.ctx, "Failed to record migration: %v", err)
		return
	}

	dbMigration := types.DBMigration{
		Migration: migration,
		ID:        migrationID,
		Schema:    schema,
		Created:   graphql.Time{Time: time.Now()},
	}
	version.DBMigrations = append(version.DBMigrations, dbMigration)
}

func (mc *mongoDBConnector) getNextSequence(name string) int32 {
	col := mc.db.Collection("counters")
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result bson.M
	err := col.FindOneAndUpdate(mc.ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 1
	}

	return result["seq"].(int32)
}

func (mc *mongoDBConnector) docToDBMigration(doc bson.M) types.DBMigration {
	return types.DBMigration{
		Migration: types.Migration{
			Name:          doc["name"].(string),
			SourceDir:     doc["source_dir"].(string),
			File:          doc["filename"].(string),
			MigrationType: types.MigrationType(doc["type"].(int32)),
			Contents:      doc["contents"].(string),
			CheckSum:      doc["checksum"].(string),
		},
		ID:      doc["_id"].(int32),
		Schema:  doc["db_schema"].(string),
		Created: graphql.Time{Time: mc.convertToTime(doc["created"])},
	}
}

func (mc *mongoDBConnector) computeSummary(summary *types.Summary, migrations []types.Migration, tenants []types.Tenant) {
	for _, migration := range migrations {
		switch migration.MigrationType {
		case types.MigrationTypeSingleMigration:
			summary.SingleMigrations++
		case types.MigrationTypeSingleScript:
			summary.SingleScripts++
		case types.MigrationTypeTenantMigration:
			summary.TenantMigrations++
		case types.MigrationTypeTenantScript:
			summary.TenantScripts++
		}
	}
	summary.TenantMigrationsTotal = summary.TenantMigrations * int32(len(tenants))
	summary.TenantScriptsTotal = summary.TenantScripts * int32(len(tenants))
	summary.MigrationsGrandTotal = summary.SingleMigrations + summary.TenantMigrationsTotal
	summary.ScriptsGrandTotal = summary.SingleScripts + summary.TenantScriptsTotal
}
