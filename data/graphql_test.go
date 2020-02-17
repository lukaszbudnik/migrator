package data

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	graphql "github.com/graph-gophers/graphql-go"
)

func TestSourceMigrationsNoFilters(t *testing.T) {

	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(schemaString, &RootResolver{loader: &mockedLoader{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations {
      sourceMigrations {
        name,
        migrationType,
        sourceDir,
      	file,
      	contents,
        checkSum
      }
    }`
	variables := map[string]interface{}{}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["sourceMigrations"].([]interface{}))
	assert.Equal(t, 5, results)
}

func TestSourceMigrationsTypeFilter(t *testing.T) {

	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(schemaString, &RootResolver{loader: &mockedLoader{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($migrationType: MigrationTypeOptional) {
	    sourceMigrations(migrationType: $migrationType) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"migrationType": 1,
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["sourceMigrations"].([]interface{}))
	assert.Equal(t, 4, results)
}

func TestSourceMigrationsTypeSourceDirFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(schemaString, &RootResolver{loader: &mockedLoader{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($sourceDir: String, $migrationType: MigrationTypeOptional) {
	    sourceMigrations(sourceDir: $sourceDir, migrationType: $migrationType) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"migrationType": 1,
		"sourceDir":     "source",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["sourceMigrations"].([]interface{}))
	assert.Equal(t, 3, results)
}

func TestSourceMigrationsTypeSourceDirNameFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(schemaString, &RootResolver{loader: &mockedLoader{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($name: String, $migrationType: MigrationTypeOptional) {
	    sourceMigrations(name: $name, migrationType: $migrationType) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"migrationType": 1,
		"name":          "201602220001.sql",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["sourceMigrations"].([]interface{}))
	assert.Equal(t, 2, results)
}

func TestSourceMigrationsTypeNameFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(schemaString, &RootResolver{loader: &mockedLoader{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($file: String) {
	    sourceMigrations(file: $file) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"file": "config/201602220001.sql",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["sourceMigrations"].([]interface{}))
	assert.Equal(t, 1, results)
}

func TestSourceMigration(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(schemaString, &RootResolver{loader: &mockedLoader{}}, opts...)

	opName := "SourceMigration"
	query := `query SourceMigration($file: String!) {
	    sourceMigration(file: $file) {
	      name,
	      migrationType,
	      sourceDir,
	    	file
	    }
  }`
	variables := map[string]interface{}{
		"file": "config/201602220001.sql",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := jsonMap["sourceMigration"].(map[string]interface{})
	assert.Equal(t, "config", results["sourceDir"])
	assert.Equal(t, "201602220001.sql", results["name"])
	assert.Equal(t, "config/201602220001.sql", results["file"])
}

func TestSourceMigrations2Queries(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(schemaString, &RootResolver{loader: &mockedLoader{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($singleMigrationType: MigrationTypeOptional, $tenantMigrationType: MigrationTypeOptional) {
	    singleTenantMigrations: sourceMigrations(migrationType: $singleMigrationType) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
      multiTenantMigrations: sourceMigrations(migrationType: $tenantMigrationType) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"singleMigrationType": 1,
		"tenantMigrationType": 2,
	}

	resp := schema.Exec(ctx, query, opName, variables)

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	single := len(jsonMap["singleTenantMigrations"].([]interface{}))
	assert.Equal(t, 4, single)
	multi := len(jsonMap["multiTenantMigrations"].([]interface{}))
	assert.Equal(t, 1, multi)
}
