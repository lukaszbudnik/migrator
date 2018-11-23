package migrations

import (
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMigrationsFlattenMigrationDBs1(t *testing.T) {

	m1 := types.MigrationDefinition{"001.sql", "public", "public/001.sql", types.MigrationTypeSingleSchema}
	db1 := types.MigrationDB{m1, "public", time.Now()}

	m2 := types.MigrationDefinition{"002.sql", "tenants", "tenants/002.sql", types.MigrationTypeTenantSchema}
	db2 := types.MigrationDB{m2, "abc", time.Now()}

	db3 := types.MigrationDB{m2, "def", time.Now()}

	m4 := types.MigrationDefinition{"003.sql", "ref", "ref/003.sql", types.MigrationTypeSingleSchema}
	db4 := types.MigrationDB{m4, "ref", time.Now()}

	dbs := []types.MigrationDB{db1, db2, db3, db4}

	migrations := flattenMigrationDBs(dbs)

	assert.Equal(t, []types.MigrationDefinition{m1, m2, m4}, migrations)

}

func TestMigrationsFlattenMigrationDBs2(t *testing.T) {

	m2 := types.MigrationDefinition{"002.sql", "tenants", "tenants/002.sql", types.MigrationTypeTenantSchema}
	db2 := types.MigrationDB{m2, "abc", time.Now()}

	db3 := types.MigrationDB{m2, "def", time.Now()}

	m4 := types.MigrationDefinition{"003.sql", "ref", "ref/003.sql", types.MigrationTypeSingleSchema}
	db4 := types.MigrationDB{m4, "ref", time.Now()}

	dbs := []types.MigrationDB{db2, db3, db4}

	migrations := flattenMigrationDBs(dbs)

	assert.Equal(t, []types.MigrationDefinition{m2, m4}, migrations)

}

func TestComputeMigrationsToApply(t *testing.T) {
	mdef1 := types.MigrationDefinition{"a", "a", "a", types.MigrationTypeSingleSchema}
	mdef2 := types.MigrationDefinition{"b", "b", "b", types.MigrationTypeTenantSchema}
	mdef3 := types.MigrationDefinition{"c", "c", "c", types.MigrationTypeTenantSchema}
	mdef4 := types.MigrationDefinition{"d", "d", "d", types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{{mdef1, ""}, {mdef2, ""}, {mdef3, ""}, {mdef4, ""}}
	dbMigrations := []types.MigrationDB{{mdef1, "a", time.Now()}, {mdef2, "abc", time.Now()}, {mdef2, "def", time.Now()}}
	migrations := ComputeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 2)

	assert.Equal(t, "c", migrations[0].File)
	assert.Equal(t, "d", migrations[1].File)
}

func TestComputeMigrationsToApplyDifferentTimestamps(t *testing.T) {
	// use case:
	// development done in parallel, 2 devs fork from master
	// dev1 adds migrations on Monday
	// dev2 adds migrations on Tuesday
	// dev2 merges and deploys his code on Tuesday
	// dev1 merges and deploys his code on Wednesday
	// migrator should detect dev1 migrations
	// previous implementation relied only on counts and such migration was not applied

	// todo: add some public migrations too
	mdef1 := types.MigrationDefinition{"20181111", "tenants", "tenants/20181111", types.MigrationTypeTenantSchema}
	mdef2 := types.MigrationDefinition{"20181111", "public", "public/20181111", types.MigrationTypeSingleSchema}
	mdef3 := types.MigrationDefinition{"20181112", "public", "public/20181112", types.MigrationTypeSingleSchema}

	dev1 := types.MigrationDefinition{"20181119", "tenants", "tenants/20181119", types.MigrationTypeTenantSchema}
	dev1p1 := types.MigrationDefinition{"201811190", "public", "public/201811190", types.MigrationTypeSingleSchema}
	dev1p2 := types.MigrationDefinition{"20181191", "public", "public/201811191", types.MigrationTypeSingleSchema}

	dev2 := types.MigrationDefinition{"20181120", "tenants", "tenants/20181120", types.MigrationTypeTenantSchema}
	dev2p := types.MigrationDefinition{"20181120", "public", "public/20181120", types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{{mdef1, ""}, {mdef2, ""}, {mdef3, ""}, {dev1, ""}, {dev1p1, ""}, {dev1p2, ""}, {dev2, ""}, {dev2p, ""}}
	dbMigrations := []types.MigrationDB{{mdef1, "abc", time.Now()}, {mdef1, "def", time.Now()}, {mdef2, "public", time.Now()}, {mdef3, "public", time.Now()}, {dev2, "abc", time.Now()}, {dev2, "def", time.Now()}, {dev2p, "public", time.Now()}}
	migrations := ComputeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 3)

	assert.Equal(t, dev1.File, migrations[0].File)
	assert.Equal(t, dev1p1.File, migrations[1].File)
	assert.Equal(t, dev1p2.File, migrations[2].File)
}
