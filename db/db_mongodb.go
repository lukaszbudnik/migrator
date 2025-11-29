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

	// Support custom tenant collection via config
	collectionName := mc.getTenantCollectionName()
	col := mc.db.Collection(collectionName)

	// Support custom field name via config
	fieldName := mc.getTenantFieldName()

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
		if name, ok := doc[fieldName].(string); ok {
			tenants = append(tenants, types.Tenant{Name: name})
		}
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
			// Use source directory as database name (consistent with SQL implementations)
			dbName := migration.SourceDir
			if action == types.ActionApply {
				mc.executeMigration(migration, dbName)
			}
			mc.recordMigration(versionID, migration, dbName, version)
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
	collectionName := mc.getTenantCollectionName()
	fieldName := mc.getTenantFieldName()
	tenantsCol := mc.db.Collection(collectionName)
	tenantDoc := bson.M{fieldName: tenantName, "created": time.Now()}
	_, err := tenantsCol.InsertOne(mc.ctx, tenantDoc)
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

func (mc *mongoDBConnector) getTenantCollectionName() string {
	// Check if custom tenant select is configured
	// Format: "collection_name" or "collection_name.field_name"
	tenantSelect := mc.config.GetTenantSelect()
	if tenantSelect != "" {
		// Log warning if using deprecated field
		if mc.config.IsUsingDeprecatedTenantSelectSQL() {
			common.LogWarn(mc.ctx, "tenantSelectSQL is deprecated since v2025.1.0 and will be removed in v2027.0.0. Use tenantSelect instead.")
		}
		// Parse collection name (before dot if present)
		parts := strings.Split(tenantSelect, ".")
		return parts[0]
	}
	return migratorTenantsTable
}

func (mc *mongoDBConnector) getTenantFieldName() string {
	// Check if custom tenant select is configured
	// Format: "collection_name" or "collection_name.field_name"
	tenantSelect := mc.config.GetTenantSelect()
	if tenantSelect != "" {
		parts := strings.Split(tenantSelect, ".")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return "name"
}

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

	// Parse and execute JavaScript-like MongoDB commands
	// This handles common patterns like db.collection.insertOne(), db.collection.createIndex(), etc.
	lines := strings.Split(contents, ";")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if err := mc.executeMongoDBCommand(targetDB, line); err != nil {
			common.LogError(mc.ctx, "Failed to execute command in migration %s (database %s): %v", migration.File, dbName, err)
		}
	}
}

// executeMongoDBCommand parses and executes a MongoDB command
func (mc *mongoDBConnector) executeMongoDBCommand(targetDB *mongo.Database, command string) error {
	command = strings.TrimSpace(command)

	// Match pattern: db.collectionName.operation(...)
	if !strings.HasPrefix(command, "db.") {
		return nil // Skip non-db commands
	}

	// Extract collection name and operation
	parts := strings.SplitN(command[3:], ".", 2) // Remove "db." prefix
	if len(parts) < 2 {
		return fmt.Errorf("invalid command format: %s", command)
	}

	collectionName := parts[0]
	rest := parts[1]

	// Extract operation name and arguments
	opEnd := strings.Index(rest, "(")
	if opEnd == -1 {
		return fmt.Errorf("no operation found in: %s", command)
	}

	operation := rest[:opEnd]
	col := targetDB.Collection(collectionName)

	// Handle different operations
	switch operation {
	case "insertOne":
		return mc.handleInsertOne(col, rest[opEnd:])
	case "createIndex":
		return mc.handleCreateIndex(col, rest[opEnd:])
	default:
		common.LogWarn(mc.ctx, "Unsupported operation: %s", operation)
		return nil
	}
}

// handleInsertOne executes insertOne operation
func (mc *mongoDBConnector) handleInsertOne(col *mongo.Collection, args string) error {
	// Extract JSON document from insertOne({...})
	start := strings.Index(args, "{")
	end := strings.LastIndex(args, "}")
	if start == -1 || end == -1 {
		return fmt.Errorf("invalid insertOne syntax")
	}

	jsonDoc := args[start : end+1]

	// Convert JavaScript object notation to proper JSON
	jsonDoc = mc.jsToJSON(jsonDoc)

	// Parse JSON to BSON
	var doc bson.M
	if err := bson.UnmarshalExtJSON([]byte(jsonDoc), false, &doc); err != nil {
		return fmt.Errorf("failed to parse document: %v", err)
	}

	_, err := col.InsertOne(mc.ctx, doc)
	return err
}

// handleCreateIndex executes createIndex operation
func (mc *mongoDBConnector) handleCreateIndex(col *mongo.Collection, args string) error {
	// Extract index spec and options from createIndex({...}, {...})
	start := strings.Index(args, "{")
	if start == -1 {
		return fmt.Errorf("invalid createIndex syntax")
	}

	// Find the matching closing brace for the first argument
	braceCount := 0
	firstArgEnd := -1
	for i := start; i < len(args); i++ {
		if args[i] == '{' {
			braceCount++
		} else if args[i] == '}' {
			braceCount--
			if braceCount == 0 {
				firstArgEnd = i
				break
			}
		}
	}

	if firstArgEnd == -1 {
		return fmt.Errorf("invalid createIndex syntax")
	}

	indexSpec := args[start : firstArgEnd+1]

	// Convert JavaScript to JSON
	indexSpec = mc.jsToJSON(indexSpec)

	// Parse index specification
	var keys bson.D
	if err := bson.UnmarshalExtJSON([]byte(indexSpec), false, &keys); err != nil {
		return fmt.Errorf("failed to parse index spec: %v", err)
	}

	// Check for options (second argument)
	opts := options.Index()
	remaining := strings.TrimSpace(args[firstArgEnd+1:])
	if strings.HasPrefix(remaining, ",") {
		remaining = strings.TrimSpace(remaining[1:])
		optStart := strings.Index(remaining, "{")
		optEnd := strings.LastIndex(remaining, "}")
		if optStart != -1 && optEnd != -1 {
			optJSON := remaining[optStart : optEnd+1]
			optJSON = mc.jsToJSON(optJSON)
			var optDoc bson.M
			if err := bson.UnmarshalExtJSON([]byte(optJSON), false, &optDoc); err == nil {
				if unique, ok := optDoc["unique"].(bool); ok {
					opts.SetUnique(unique)
				}
			}
		}
	}

	_, err := col.Indexes().CreateOne(mc.ctx, mongo.IndexModel{
		Keys:    keys,
		Options: opts,
	})
	return err
}

// jsToJSON converts JavaScript object notation to proper JSON
// Handles: {key: value} -> {"key": value}, {key: 'value'} -> {"key": "value"}
func (mc *mongoDBConnector) jsToJSON(js string) string {
	// Replace single quotes with double quotes
	result := strings.ReplaceAll(js, "'", "\"")

	// Add quotes around unquoted keys
	// Match pattern: {word: or ,word: and replace with {"word":
	result = strings.ReplaceAll(result, "{", "{ ")
	result = strings.ReplaceAll(result, ",", ", ")

	// Simple regex-like replacement for unquoted keys
	var builder strings.Builder
	inQuotes := false
	i := 0
	for i < len(result) {
		ch := result[i]

		if ch == '"' {
			inQuotes = !inQuotes
			builder.WriteByte(ch)
			i++
			continue
		}

		if !inQuotes && (ch == '{' || ch == ',' || ch == ' ') {
			builder.WriteByte(ch)
			i++
			// Skip whitespace
			for i < len(result) && result[i] == ' ' {
				builder.WriteByte(result[i])
				i++
			}
			// Check if next is an unquoted key
			if i < len(result) && result[i] != '"' && result[i] != '}' {
				// Find the key
				keyStart := i
				for i < len(result) && result[i] != ':' && result[i] != ' ' {
					i++
				}
				if i < len(result) && result[i] == ':' || (i < len(result)-1 && result[i] == ' ' && result[i+1] == ':') {
					key := result[keyStart:i]
					key = strings.TrimSpace(key)
					if key != "" && key[0] != '"' {
						builder.WriteByte('"')
						builder.WriteString(key)
						builder.WriteByte('"')
						continue
					}
				}
				// Not a key, write as-is
				builder.WriteString(result[keyStart:i])
			}
		} else {
			builder.WriteByte(ch)
			i++
		}
	}

	return builder.String()
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
	col := mc.db.Collection("migrator_counters")
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
