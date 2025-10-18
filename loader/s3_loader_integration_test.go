package loader

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestS3GetSourceMigrationsIntegration(t *testing.T) {
	bucketName := os.Getenv("S3_BUCKET")

	if len(bucketName) == 0 {
		t.Skip("skipping integration test: S3_BUCKET not set")
	}

	baseLocation := fmt.Sprintf("s3://%s", bucketName)

	config := &config.Config{
		BaseLocation:     baseLocation,
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &s3Loader{
		baseLoader:       baseLoader{context.TODO(), config},
		clientFactory:    &defaultS3ClientFactory{},
		paginatorFactory: &defaultS3PaginatorFactory{},
	}

	migrations := loader.GetSourceMigrations()

	assert.Len(t, migrations, 16)

	assert.Contains(t, migrations[0].File, "migrations/tenants-scripts/200001181228.sql")
	assert.Contains(t, migrations[1].File, "migrations/config-scripts/200012181227.sql")
	assert.Contains(t, migrations[2].File, "migrations/config/201602160001.sql")
	assert.Contains(t, migrations[3].File, "migrations/config/201602160002.sql")
	assert.Contains(t, migrations[4].File, "migrations/tenants/201602160002.sql")
	assert.Contains(t, migrations[5].File, "migrations/ref/201602160003.sql")
	assert.Contains(t, migrations[6].File, "migrations/tenants/201602160003.sql")
	assert.Contains(t, migrations[7].File, "migrations/ref/201602160004.sql")
	assert.Contains(t, migrations[8].File, "migrations/tenants/201602160004.sql")
	assert.Contains(t, migrations[9].File, "migrations/tenants/201602160005.sql")
	assert.Contains(t, migrations[10].File, "migrations/tenants-scripts/a.sql")
	assert.Contains(t, migrations[11].File, "migrations/tenants-scripts/b.sql")
	assert.Contains(t, migrations[12].File, "migrations/config-scripts/200012181227.sql")
	assert.Contains(t, migrations[13].File, "migrations/tenants-scripts/200001181228.sql")
	assert.Contains(t, migrations[14].File, "migrations/tenants-scripts/a.sql")
	assert.Contains(t, migrations[15].File, "migrations/tenants-scripts/b.sql")
}

func TestS3GetSourceMigrationsBucketWithPrefixIntegration(t *testing.T) {
	bucketName := os.Getenv("S3_BUCKET")

	if len(bucketName) == 0 {
		t.Skip("skipping integration test: S3_BUCKET not set")
	}

	baseLocation := fmt.Sprintf("s3://%s/app-x/", bucketName)

	config := &config.Config{
		BaseLocation:     baseLocation,
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &s3Loader{
		baseLoader:       baseLoader{context.TODO(), config},
		clientFactory:    &defaultS3ClientFactory{},
		paginatorFactory: &defaultS3PaginatorFactory{},
	}

	migrations := loader.GetSourceMigrations()

	assert.Len(t, migrations, 16)

	assert.Contains(t, migrations[0].File, "app-x/migrations/tenants-scripts/200001181228.sql")
	assert.Contains(t, migrations[1].File, "app-x/migrations/config-scripts/200012181227.sql")
	assert.Contains(t, migrations[2].File, "app-x/migrations/config/201602160001.sql")
	assert.Contains(t, migrations[3].File, "app-x/migrations/config/201602160002.sql")
	assert.Contains(t, migrations[4].File, "app-x/migrations/tenants/201602160002.sql")
	assert.Contains(t, migrations[5].File, "app-x/migrations/ref/201602160003.sql")
	assert.Contains(t, migrations[6].File, "app-x/migrations/tenants/201602160003.sql")
	assert.Contains(t, migrations[7].File, "app-x/migrations/ref/201602160004.sql")
	assert.Contains(t, migrations[8].File, "app-x/migrations/tenants/201602160004.sql")
	assert.Contains(t, migrations[9].File, "app-x/migrations/tenants/201602160005.sql")
	assert.Contains(t, migrations[10].File, "app-x/migrations/tenants-scripts/a.sql")
	assert.Contains(t, migrations[11].File, "app-x/migrations/tenants-scripts/b.sql")
	assert.Contains(t, migrations[12].File, "app-x/migrations/config-scripts/200012181227.sql")
	assert.Contains(t, migrations[13].File, "app-x/migrations/tenants-scripts/200001181228.sql")
	assert.Contains(t, migrations[14].File, "app-x/migrations/tenants-scripts/a.sql")
	assert.Contains(t, migrations[15].File, "app-x/migrations/tenants-scripts/b.sql")
}

func TestS3HealthCheckIntegration(t *testing.T) {
	bucketName := os.Getenv("S3_BUCKET")

	if len(bucketName) == 0 {
		t.Skip("skipping integration test: S3_BUCKET not set")
	}

	baseLocation := fmt.Sprintf("s3://%s", bucketName)

	config := &config.Config{
		BaseLocation:     baseLocation,
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &s3Loader{
		baseLoader:       baseLoader{context.TODO(), config},
		clientFactory:    &defaultS3ClientFactory{},
		paginatorFactory: &defaultS3PaginatorFactory{},
	}

	err := loader.HealthCheck()
	assert.Nil(t, err)
}
