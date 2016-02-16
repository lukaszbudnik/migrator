package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadAllMigrationsNonExistingSourceDir(t *testing.T) {

	var config Config
	config.SourceDir = "xyzabc"
	migrations, err := listAllMigrations(config)

	assert.Nil(t, migrations)
	assert.Error(t, err)

}

func TestListAllMigrationsExistingSourceDir(t *testing.T) {

	var config Config
	config.SourceDir = "test/migrations"
	config.SingleSchemas = []string{"public"}
	config.TenantSchemas = []string{"tenants"}
	migrations, err := listAllMigrations(config)

	assert.Nil(t, err)

	assert.Len(t, migrations, 6)

	assert.Equal(t, "test/migrations/public/201602160001.sql", migrations[0])
	assert.Equal(t, "test/migrations/public/201602160002.sql", migrations[1])
	assert.Equal(t, "test/migrations/tenants/201602160002.sql", migrations[2])
	assert.Equal(t, "test/migrations/tenants/201602160003.sql", migrations[3])
	assert.Equal(t, "test/migrations/public/201602160004.sql", migrations[4])
	assert.Equal(t, "test/migrations/tenants/201602160004.sql", migrations[5])

}

func TestComputeMigrationsToApply(t *testing.T) {
	migrations := computeMigrationsToApply([]string{"a", "b", "c", "d"}, []string{"a", "b"})

	assert.Len(t, migrations, 2)

	assert.Equal(t, "c", migrations[0])
	assert.Equal(t, "d", migrations[1])
}
