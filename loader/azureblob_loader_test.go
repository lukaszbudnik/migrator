package loader

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestAzureGetSourceMigrations(t *testing.T) {

	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	if len(accountName) == 0 || len(accountKey) == 0 {
		t.Skip("skipping test AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY not set")
	}

	// migrator implements env variable substitution and normally we would use:
	// "https://${AZURE_STORAGE_ACCOUNT}.blob.core.windows.net/mycontainer"
	// however below we are creating the Config struct directly
	// and that's why we need to build correct URL ourselves
	baseLocation := fmt.Sprintf("https://%v.blob.core.windows.net/mycontainer", accountName)

	config := &config.Config{
		BaseLocation:     baseLocation,
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &azureBlobLoader{baseLoader{context.TODO(), config}}
	migrations := loader.GetSourceMigrations()

	assert.Len(t, migrations, 12)

	assert.Contains(t, migrations[0].File, "migrations/config/201602160001.sql")
	assert.Contains(t, migrations[1].File, "migrations/config/201602160002.sql")
	assert.Contains(t, migrations[2].File, "migrations/tenants/201602160002.sql")
	assert.Contains(t, migrations[3].File, "migrations/ref/201602160003.sql")
	assert.Contains(t, migrations[4].File, "migrations/tenants/201602160003.sql")
	assert.Contains(t, migrations[5].File, "migrations/ref/201602160004.sql")
	assert.Contains(t, migrations[6].File, "migrations/tenants/201602160004.sql")
	assert.Contains(t, migrations[7].File, "migrations/tenants/201602160005.sql")
	assert.Contains(t, migrations[8].File, "migrations/config-scripts/200012181227.sql")
	assert.Contains(t, migrations[9].File, "migrations/tenants-scripts/200001181228.sql")
	assert.Contains(t, migrations[10].File, "migrations/tenants-scripts/a.sql")
	assert.Contains(t, migrations[11].File, "migrations/tenants-scripts/b.sql")

}

func TestAzureGetSourceMigrationsWithOptionalPrefix(t *testing.T) {

	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	if len(accountName) == 0 || len(accountKey) == 0 {
		t.Skip("skipping test AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY not set")
	}

	baseLocation := fmt.Sprintf("https://%v.blob.core.windows.net/myothercontainer/prod/artefacts/", accountName)

	config := &config.Config{
		BaseLocation:     baseLocation,
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &azureBlobLoader{baseLoader{context.TODO(), config}}
	migrations := loader.GetSourceMigrations()

	assert.Len(t, migrations, 12)

	assert.Contains(t, migrations[0].File, "prod/artefacts/migrations/config/201602160001.sql")
	assert.Contains(t, migrations[1].File, "prod/artefacts/migrations/config/201602160002.sql")
	assert.Contains(t, migrations[2].File, "prod/artefacts/migrations/tenants/201602160002.sql")
	assert.Contains(t, migrations[3].File, "prod/artefacts/migrations/ref/201602160003.sql")
	assert.Contains(t, migrations[4].File, "prod/artefacts/migrations/tenants/201602160003.sql")
	assert.Contains(t, migrations[5].File, "prod/artefacts/migrations/ref/201602160004.sql")
	assert.Contains(t, migrations[6].File, "prod/artefacts/migrations/tenants/201602160004.sql")
	assert.Contains(t, migrations[7].File, "prod/artefacts/migrations/tenants/201602160005.sql")
	assert.Contains(t, migrations[8].File, "prod/artefacts/migrations/config-scripts/200012181227.sql")
	assert.Contains(t, migrations[9].File, "prod/artefacts/migrations/tenants-scripts/200001181228.sql")
	assert.Contains(t, migrations[10].File, "prod/artefacts/migrations/tenants-scripts/a.sql")
	assert.Contains(t, migrations[11].File, "prod/artefacts/migrations/tenants-scripts/b.sql")

}
