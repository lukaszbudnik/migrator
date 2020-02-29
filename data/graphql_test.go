package data

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/graph-gophers/graphql-go"
)

func TestTenants(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "Tenants"
	query := `query Tenants {
      tenants {
        name
      }
    }`
	variables := map[string]interface{}{}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["tenants"].([]interface{}))
	assert.Equal(t, 3, results)
}

func TestVersions(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "Versions"
	query := `query Versions {
      versions {
        id
        name
        created
      }
    }`
	variables := map[string]interface{}{}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	versions := jsonMap["versions"].([]interface{})
	results := len(versions)

	assert.Equal(t, 3, results)
	assert.Equal(t, "a", versions[0].(map[string]interface{})["name"])
	assert.Equal(t, "bb", versions[1].(map[string]interface{})["name"])
	assert.Equal(t, "ccc", versions[2].(map[string]interface{})["name"])
}

func TestVersionsByFile(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "Versions"
	query := `query Versions($file: String) {
      versions(file: $file) {
        id
        name
        created
      }
    }`
	variables := map[string]interface{}{
		"file": "config/202002180000.sql",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	versions := jsonMap["versions"].([]interface{})
	results := len(versions)

	assert.Equal(t, 1, results)
	assert.Equal(t, "a", versions[0].(map[string]interface{})["name"])
}

func TestVersionByID(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "Version"
	query := `query Version($id: Int!) {
      version(id: $id) {
        id
        created
        dbMigrations {
          file
          schema
          migrationType
        }
      }
    }`
	variables := map[string]interface{}{
		"id": 1234,
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	version := jsonMap["version"].(map[string]interface{})
	dbMigrations := version["dbMigrations"].([]interface{})
	// json.Unmarshal actually creates float64 for all number, but this is only unit test
	assert.Equal(t, float64(1234), version["id"])
	assert.Nil(t, version["name"])
	assert.NotNil(t, version["created"])
	assert.Equal(t, 5, len(dbMigrations))
	lastDBMigration := dbMigrations[4].(map[string]interface{})
	assert.Equal(t, "tenants/202002180000.sql", lastDBMigration["file"])
	assert.Equal(t, "TenantMigration", lastDBMigration["migrationType"])
	assert.Equal(t, "xyz", lastDBMigration["schema"])
}

func TestSourceMigrationsNoFilters(t *testing.T) {

	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

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
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($filters: SourceMigrationFilters) {
	    sourceMigrations(filters: $filters) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"filters": map[string]interface{}{
			"migrationType": "SingleMigration",
		},
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := jsonMap["sourceMigrations"].([]interface{})
	migration := results[0].(map[string]interface{})
	assert.Equal(t, "SingleMigration", migration["migrationType"])
}

func TestSourceMigrationsTypeSourceDirFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($filters: SourceMigrationFilters) {
	    sourceMigrations(filters: $filters) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"filters": map[string]interface{}{
			"migrationType": "SingleMigration",
			"sourceDir":     "source",
		},
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := jsonMap["sourceMigrations"].([]interface{})
	migration := results[0].(map[string]interface{})
	assert.Equal(t, "SingleMigration", migration["migrationType"])
	assert.Equal(t, "source", migration["sourceDir"])
}

func TestSourceMigrationsTypeSourceDirNameFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($filters: SourceMigrationFilters) {
	    sourceMigrations(filters: $filters) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"filters": map[string]interface{}{
			"migrationType": "SingleMigration",
			"name":          "201602220001.sql",
		},
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := jsonMap["sourceMigrations"].([]interface{})
	migration := results[0].(map[string]interface{})
	assert.Equal(t, "SingleMigration", migration["migrationType"])
	assert.Equal(t, "201602220001.sql", migration["name"])
}

func TestSourceMigrationsTypeNameFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($filters: SourceMigrationFilters) {
	    sourceMigrations(filters: $filters) {
	      name,
	      migrationType,
	      sourceDir,
	    	file,
	    	contents,
	      checkSum
	    }
  }`
	variables := map[string]interface{}{
		"filters": map[string]interface{}{
			"file": "config/201602220001.sql",
		},
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
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

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
	assert.Equal(t, "config/201602220001.sql", results["file"])
	assert.NotNil(t, "201602220001.sql", results["name"])
	assert.NotNil(t, "SingleMigration", results["migrationType"])
	assert.NotNil(t, "config", results["sourceDir"])
	// we return only 4 fields in above query others should be nil
	assert.Nil(t, results["contents"])
	assert.Nil(t, results["checkSum"])
}

func TestDBMigration(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "DBMigration"
	query := `query DBMigration($id: Int!) {
	    dbMigration(id: $id) {
        name,
        migrationType,
        sourceDir,
      	file,
        checkSum,
        schema,
      	created
	    }
  }`
	variables := map[string]interface{}{
		"id": 123,
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := jsonMap["dbMigration"].(map[string]interface{})
	assert.Equal(t, "201602220000.sql", results["name"])
	assert.Equal(t, "SingleMigration", results["migrationType"])
	assert.Equal(t, "source", results["sourceDir"])
	assert.Equal(t, "source/201602220000.sql", results["file"])
	assert.Equal(t, "source", results["schema"])
	assert.Equal(t, "2016-02-22T16:41:01.000000123Z", results["created"])
	// we return all fields except contents - should be nil
	assert.Nil(t, results["contents"])
}

func TestComplexQueries(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "Data"
	query := `
  query Data($singleMigrationsFilters: SourceMigrationFilters, $tenantMigrationsFilters: SourceMigrationFilters) {
    singleTenantSourceMigrations: sourceMigrations(filters: $singleMigrationsFilters) {
      file
      migrationType
    }
    multiTenantSourceMigrations: sourceMigrations(filters: $tenantMigrationsFilters) {
      file
      migrationType
      checkSum
    }
    tenants {
      name
    }
  }
  `
	variables := map[string]interface{}{
		"singleMigrationsFilters": map[string]interface{}{
			"migrationType": "SingleMigration",
		},
		"tenantMigrationsFilters": map[string]interface{}{
			"migrationType": "TenantMigration",
		},
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	single := len(jsonMap["singleTenantSourceMigrations"].([]interface{}))
	assert.Equal(t, 1, single)
	multi := len(jsonMap["multiTenantSourceMigrations"].([]interface{}))
	assert.Equal(t, 1, multi)
	tenants := len(jsonMap["tenants"].([]interface{}))
	assert.Equal(t, 3, tenants)
}
