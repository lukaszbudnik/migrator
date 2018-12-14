package loader

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestDiskReadDiskMigrationsNonExistingBaseDirError(t *testing.T) {

	var config config.Config
	config.BaseDir = "xyzabc"

	loader := CreateLoader(&config)

	_, err := loader.GetDiskMigrations()
	assert.Equal(t, "open xyzabc: no such file or directory", err.Error())
}

func TestDiskGetDiskMigrations(t *testing.T) {

	var config config.Config
	config.BaseDir = "../test/migrations"
	config.SingleSchemas = []string{"config", "ref"}
	config.TenantSchemas = []string{"tenants"}

	loader := CreateLoader(&config)
	migrations, err := loader.GetDiskMigrations()
	assert.Nil(t, err)

	assert.Len(t, migrations, 8)

	assert.Equal(t, "config/201602160001.sql", migrations[0].File)
	assert.Equal(t, "config/201602160002.sql", migrations[1].File)
	assert.Equal(t, "tenants/201602160002.sql", migrations[2].File)
	assert.Equal(t, "ref/201602160003.sql", migrations[3].File)
	assert.Equal(t, "tenants/201602160003.sql", migrations[4].File)
	assert.Equal(t, "ref/201602160004.sql", migrations[5].File)
	assert.Equal(t, "tenants/201602160004.sql", migrations[6].File)
	assert.Equal(t, "tenants/201602160005.sql", migrations[7].File)

}
