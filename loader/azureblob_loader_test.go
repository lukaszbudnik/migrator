package loader

import (
	"context"
	"os"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestAzureGetSourceMigrations(t *testing.T) {

	travis := os.Getenv("TRAVIS")
	if len(travis) > 0 {
		t.Skip("Does not work on travis due to Azure Storage Account credentials required")
	}

	config := &config.Config{
		BaseDir:          "https://lukaszbudniktest.blob.core.windows.net/mycontainer",
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
