package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadAllMigrationsNonExistingSourceDir(t *testing.T) {

	var config Config
	config.BaseDir = "xyzabc"
	migrations, err := listAllMigrations(config)

	assert.Nil(t, migrations)
	assert.Error(t, err)

}

func TestListAllMigrationsExistingSourceDir(t *testing.T) {

	var config Config
	config.BaseDir = "test/migrations"
	config.SingleSchemas = []string{"public", "ref"}
	config.TenantSchemas = []string{"tenants"}
	migrations, err := listAllMigrations(config)

	assert.Nil(t, err)

	assert.Len(t, migrations, 6)

	assert.Equal(t, "public/201602160001.sql", migrations[0].File)
	assert.Equal(t, "tenants/201602160002.sql", migrations[1].File)
	assert.Equal(t, "tenants/201602160003.sql", migrations[2].File)
	assert.Equal(t, "public/201602160004.sql", migrations[3].File)
	assert.Equal(t, "ref/201602160004.sql", migrations[4].File)
	assert.Equal(t, "tenants/201602160004.sql", migrations[5].File)

}

func TestComputeMigrationsToApply(t *testing.T) {
	migrations := computeMigrationsToApply([]MigrationDefinition{{"a", "a", "a", ModeSingleSchema}, {"b", "b", "b", ModeSingleSchema}, {"c", "c", "c", ModeSingleSchema}, {"d", "d", "d", ModeSingleSchema}}, []MigrationDefinition{{"a", "a", "a", ModeSingleSchema}, {"b", "b", "b", ModeSingleSchema}})

	assert.Len(t, migrations, 2)

	assert.Equal(t, "c", migrations[0].File)
	assert.Equal(t, "d", migrations[1].File)
}
