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
	query := `query SourceMigrations($migrationType: MigrationType) {
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
		"migrationType": "SingleMigration",
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
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($sourceDir: String, $migrationType: MigrationType) {
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
		"migrationType": "SingleMigration",
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
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "SourceMigrations"
	query := `query SourceMigrations($name: String, $migrationType: MigrationType) {
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
		"migrationType": "SingleMigration",
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
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

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
	assert.Equal(t, "201602220001.sql", results["name"])
	assert.Equal(t, "SingleMigration", results["migrationType"])
	assert.Equal(t, "config", results["sourceDir"])
	assert.Equal(t, "config/201602220001.sql", results["file"])
	// we return only 4 fields in above query others should be nil
	assert.Nil(t, results["contents"])
}

func TestDBMigrationsNoFilters(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "DBMigrations"
	query := `query DBMigrations {
      dbMigrations {
        migrationType,
      	file,
        schema
      }
    }`
	variables := map[string]interface{}{}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["dbMigrations"].([]interface{}))
	assert.Equal(t, 5, results)
}

// migrationType is defined at the Migration level
func TestDBMigrationsMigrationTypeFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "DBMigrations"
	query := `query DBMigrations($migrationType: MigrationType) {
	    dbMigrations(migrationType: $migrationType) {
	      name,
	      migrationType,
	      sourceDir,
	    	file
	    }
  }`
	variables := map[string]interface{}{
		"migrationType": "SingleMigration",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["dbMigrations"].([]interface{}))
	assert.Equal(t, 2, results)
}

// migrationType is defined at the Migration level
// schema is defined at the MigrationDB level
func TestDBMigrationsSchemaMigrationTypeFilter(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "DBMigrations"
	query := `query DBMigrations($schema: String, $migrationType: MigrationType) {
	    dbMigrations(schema: $schema, migrationType: $migrationType) {
	      name,
	      migrationType,
	      sourceDir,
	    	file
	    }
  }`
	variables := map[string]interface{}{
		"migrationType": "TenantMigration",
		"schema":        "abc",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := len(jsonMap["dbMigrations"].([]interface{}))
	assert.Equal(t, 1, results)
}

func TestDBMigration(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "DBMigration"
	query := `query DBMigration($schema: String!, $file: String!) {
	    dbMigration(schema: $schema, file: $file) {
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
		"file":   "tenants/202002180000.sql",
		"schema": "xyz",
	}

	resp := schema.Exec(ctx, query, opName, variables)
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	results := jsonMap["dbMigration"].(map[string]interface{})
	assert.Equal(t, "202002180000.sql", results["name"])
	assert.Equal(t, "TenantMigration", results["migrationType"])
	assert.Equal(t, "tenants", results["sourceDir"])
	assert.Equal(t, "tenants/202002180000.sql", results["file"])
	assert.Equal(t, "xyz", results["schema"])
	assert.Equal(t, "2020-02-18T16:41:01.000000123Z", results["created"])
	// we return all fields except contents - should be nil
	assert.Nil(t, results["contents"])
}

func TestComplexQueries(t *testing.T) {
	ctx := context.Background()

	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(SchemaDefinition, &RootResolver{Coordinator: &mockedCoordinator{}}, opts...)

	opName := "Data"
	query := `
  query Data($singleMigrationType: MigrationType, $tenantMigrationType: MigrationType) {
    singleTenantSourceMigrations: sourceMigrations(migrationType: $singleMigrationType) {
      file
      migrationType
    }
    multiTenantDBMigrations: dbMigrations(migrationType: $tenantMigrationType) {
      file
      migrationType
      schema
      checkSum
      created
    }
    tenants {
      name
    }
  }
  `
	variables := map[string]interface{}{
		"singleMigrationType": "SingleMigration",
		"tenantMigrationType": "TenantMigration",
	}

	resp := schema.Exec(ctx, query, opName, variables)

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(resp.Data, &jsonMap)
	assert.Nil(t, err)
	single := len(jsonMap["singleTenantSourceMigrations"].([]interface{}))
	assert.Equal(t, 4, single)
	multi := len(jsonMap["multiTenantDBMigrations"].([]interface{}))
	assert.Equal(t, 3, multi)
	tenants := len(jsonMap["tenants"].([]interface{}))
	assert.Equal(t, 3, tenants)
}
