package loader

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiskPanicReadDiskMigrationsNonExistingBaseDir(t *testing.T) {

	var config config.Config
	config.BaseDir = "xyzabc"

	loader := CreateLoader(&config)

	assert.Panics(t, func() {
		loader.GetMigrations()
	}, "Should panic because of non-existing base dir")

}

func TestDiskGetDiskMigrations(t *testing.T) {

	var config config.Config
	config.BaseDir = "../test/migrations"
	config.SingleSchemas = []string{"public", "ref"}
	config.TenantSchemas = []string{"tenants"}

	loader := CreateLoader(&config)
	migrations := loader.GetMigrations()

	assert.Len(t, migrations, 6)

	assert.Equal(t, "public/201602160001.sql", migrations[0].File)
	assert.Equal(t, "tenants/201602160002.sql", migrations[1].File)
	assert.Equal(t, "tenants/201602160003.sql", migrations[2].File)
	assert.Equal(t, "public/201602160004.sql", migrations[3].File)
	assert.Equal(t, "ref/201602160004.sql", migrations[4].File)
	assert.Equal(t, "tenants/201602160004.sql", migrations[5].File)

}
