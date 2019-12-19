package loader

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestDiskReadDiskMigrationsNonExistingBaseDirError(t *testing.T) {
	var config config.Config
	config.BaseDir = "xyzabc"

	loader := NewLoader(&config)

	_, err := loader.GetDiskMigrations()
	assert.Equal(t, "open xyzabc: no such file or directory", err.Error())
}

func TestDiskGetDiskMigrations(t *testing.T) {
	var config config.Config
	config.BaseDir = "../test/migrations"
	config.SingleMigrations = []string{"config", "ref"}
	config.TenantMigrations = []string{"tenants"}
	config.SingleScripts = []string{"config-scripts"}
	config.TenantScripts = []string{"tenants-scripts"}

	loader := NewLoader(&config)
	migrations, err := loader.GetDiskMigrations()
	assert.Nil(t, err)

	assert.Len(t, migrations, 10)

	assert.Equal(t, "config/201602160001.sql", migrations[0].File)
	assert.Equal(t, "config/201602160002.sql", migrations[1].File)
	assert.Equal(t, "tenants/201602160002.sql", migrations[2].File)
	assert.Equal(t, "ref/201602160003.sql", migrations[3].File)
	assert.Equal(t, "tenants/201602160003.sql", migrations[4].File)
	assert.Equal(t, "ref/201602160004.sql", migrations[5].File)
	assert.Equal(t, "tenants/201602160004.sql", migrations[6].File)
	assert.Equal(t, "tenants/201602160005.sql", migrations[7].File)
	assert.Equal(t, "config-scripts/201912181227.sql", migrations[8].File)
	assert.Equal(t, "tenants-scripts/201912181228.sql", migrations[9].File)
}
