package migrations

import (
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func TestMigrationsFlattenMigrationDBs1(t *testing.T) {
	m1 := types.Migration{Name: "001.sql", SourceDir: "public", File: "public/001.sql", MigrationType: types.MigrationTypeSingleSchema}
	db1 := types.MigrationDB{Migration: m1, Schema: "public", Created: time.Now()}

	m2 := types.Migration{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantSchema}
	db2 := types.MigrationDB{Migration: m2, Schema: "abc", Created: time.Now()}

	db3 := types.MigrationDB{Migration: m2, Schema: "def", Created: time.Now()}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleSchema}
	db4 := types.MigrationDB{Migration: m4, Schema: "ref", Created: time.Now()}

	dbs := []types.MigrationDB{db1, db2, db3, db4}

	migrations := flattenMigrationDBs(dbs)

	assert.Equal(t, []types.Migration{m1, m2, m4}, migrations)

}

func TestMigrationsFlattenMigrationDBs2(t *testing.T) {

	m2 := types.Migration{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantSchema}
	db2 := types.MigrationDB{Migration: m2, Schema: "abc", Created: time.Now()}

	db3 := types.MigrationDB{Migration: m2, Schema: "def", Created: time.Now()}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleSchema}
	db4 := types.MigrationDB{Migration: m4, Schema: "ref", Created: time.Now()}

	dbs := []types.MigrationDB{db2, db3, db4}

	migrations := flattenMigrationDBs(dbs)

	assert.Equal(t, []types.Migration{m2, m4}, migrations)

}

func TestComputeMigrationsToApply(t *testing.T) {
	mdef1 := types.Migration{Name: "a", SourceDir: "a", File: "a", MigrationType: types.MigrationTypeSingleSchema}
	mdef2 := types.Migration{Name: "b", SourceDir: "b", File: "b", MigrationType: types.MigrationTypeTenantSchema}
	mdef3 := types.Migration{Name: "c", SourceDir: "c", File: "c", MigrationType: types.MigrationTypeTenantSchema}
	mdef4 := types.Migration{Name: "d", SourceDir: "d", File: "d", MigrationType: types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, mdef4}
	dbMigrations := []types.MigrationDB{{Migration: mdef1, Schema: "a", Created: time.Now()}, {Migration: mdef2, Schema: "abc", Created: time.Now()}, {Migration: mdef2, Schema: "def", Created: time.Now()}}
	migrations := ComputeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 2)

	assert.Equal(t, "c", migrations[0].File)
	assert.Equal(t, "d", migrations[1].File)
}

func TestFilterTenantMigrations(t *testing.T) {
	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantSchema}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleSchema}
	mdef3 := types.Migration{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleSchema}

	dev1 := types.Migration{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantSchema}
	dev1p1 := types.Migration{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleSchema}
	dev1p2 := types.Migration{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleSchema}

	dev2 := types.Migration{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantSchema}
	dev2p := types.Migration{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, dev1, dev1p1, dev1p2, dev2, dev2p}
	migrations := FilterTenantMigrations(diskMigrations)

	assert.Len(t, migrations, 3)

	assert.Equal(t, mdef1.File, migrations[0].File)
	assert.Equal(t, types.MigrationTypeTenantSchema, migrations[0].MigrationType)
	assert.Equal(t, dev1.File, migrations[1].File)
	assert.Equal(t, types.MigrationTypeTenantSchema, migrations[1].MigrationType)
	assert.Equal(t, dev2.File, migrations[2].File)
	assert.Equal(t, types.MigrationTypeTenantSchema, migrations[2].MigrationType)
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

	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantSchema}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleSchema}
	mdef3 := types.Migration{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleSchema}

	dev1 := types.Migration{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantSchema}
	dev1p1 := types.Migration{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleSchema}
	dev1p2 := types.Migration{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleSchema}

	dev2 := types.Migration{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantSchema}
	dev2p := types.Migration{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, dev1, dev1p1, dev1p2, dev2, dev2p}
	dbMigrations := []types.MigrationDB{{Migration: mdef1, Schema: "abc", Created: time.Now()}, {Migration: mdef1, Schema: "def", Created: time.Now()}, {Migration: mdef2, Schema: "public", Created: time.Now()}, {Migration: mdef3, Schema: "public", Created: time.Now()}, {Migration: dev2, Schema: "abc", Created: time.Now()}, {Migration: dev2, Schema: "def", Created: time.Now()}, {Migration: dev2p, Schema: "public", Created: time.Now()}}
	migrations := ComputeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 3)

	assert.Equal(t, dev1.File, migrations[0].File)
	assert.Equal(t, dev1p1.File, migrations[1].File)
	assert.Equal(t, dev1p2.File, migrations[2].File)
}

func TestIntersect(t *testing.T) {
	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantSchema}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleSchema}
	mdef3 := types.Migration{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleSchema}

	dev1 := types.Migration{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantSchema}
	dev1p1 := types.Migration{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleSchema}
	dev1p2 := types.Migration{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleSchema}

	dev2 := types.Migration{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantSchema}
	dev2p := types.Migration{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, dev1, dev1p1, dev1p2, dev2, dev2p}
	dbMigrations := []types.Migration{mdef1, mdef2, mdef3, dev2, dev2p}

	intersect := intersect(diskMigrations, dbMigrations)
	assert.Len(t, intersect, 5)
	for i := range intersect {
		assert.Equal(t, intersect[i].disk, intersect[i].db)
		assert.Equal(t, intersect[i].disk, dbMigrations[i])
	}
}

func TestVerifyCheckSumsOK(t *testing.T) {
	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantSchema, CheckSum: "abc"}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeSingleSchema, CheckSum: "abc"}
	verified, offendingMigrations := VerifyCheckSums([]types.Migration{mdef1}, []types.MigrationDB{{Migration: mdef2}})
	assert.True(t, verified)
	assert.Empty(t, offendingMigrations)
}

func TestVerifyCheckSumsKO(t *testing.T) {
	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantSchema, CheckSum: "abc"}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeSingleSchema, CheckSum: "abcd"}
	verified, offendingMigrations := VerifyCheckSums([]types.Migration{mdef1}, []types.MigrationDB{{Migration: mdef2}})
	assert.False(t, verified)
	assert.Equal(t, mdef1, offendingMigrations[0])
}
