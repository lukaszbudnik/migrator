package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReadAllMigrationsNonExistingSourceDir(t *testing.T) {

	var config Config
	config.BaseDir = "xyzabc"
	migrations, err := listAllMigrations(config)

	assert.Nil(t, migrations)
	assert.Error(t, err)

}

func TestFlattenDBMigrations1(t *testing.T) {

	m1 := MigrationDefinition{"001.sql", "public", "public/001.sql", ModeSingleSchema}
	db1 := DBMigration{m1, "public", time.Now()}

	m2 := MigrationDefinition{"002.sql", "tenants", "tenants/002.sql", ModeTenantSchema}
	db2 := DBMigration{m2, "abc", time.Now()}

	db3 := DBMigration{m2, "def", time.Now()}

	m4 := MigrationDefinition{"003.sql", "ref", "ref/003.sql", ModeSingleSchema}
	db4 := DBMigration{m4, "ref", time.Now()}

	dbs := []DBMigration{db1, db2, db3, db4}

	migrations := flattenDBMigrations(dbs)

	assert.Equal(t, []MigrationDefinition{m1, m2, m4}, migrations)

}

func TestFlattenDBMigrations2(t *testing.T) {

	m2 := MigrationDefinition{"002.sql", "tenants", "tenants/002.sql", ModeTenantSchema}
	db2 := DBMigration{m2, "abc", time.Now()}

	db3 := DBMigration{m2, "def", time.Now()}

	m4 := MigrationDefinition{"003.sql", "ref", "ref/003.sql", ModeSingleSchema}
	db4 := DBMigration{m4, "ref", time.Now()}

	dbs := []DBMigration{db2, db3, db4}

	migrations := flattenDBMigrations(dbs)

	assert.Equal(t, []MigrationDefinition{m2, m4}, migrations)

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

// func TestComputeMigrationsToApply(t *testing.T) {
// 	migrations := computeMigrationsToApply([]MigrationDefinition{{"a", "a", "a", ModeSingleSchema}, {"b", "b", "b", ModeSingleSchema}, {"c", "c", "c", ModeSingleSchema}, {"d", "d", "d", ModeSingleSchema}}, []MigrationDefinition{{"a", "a", "a", ModeSingleSchema}, {"b", "b", "b", ModeSingleSchema}})
//
// 	assert.Len(t, migrations, 2)
//
// 	assert.Equal(t, "c", migrations[0].File)
// 	assert.Equal(t, "d", migrations[1].File)
// }
