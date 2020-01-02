package coordinator

import (
	"context"
	"testing"
	"time"

	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func TestMigrationsFlattenMigrationDBs1(t *testing.T) {
	m1 := types.Migration{Name: "001.sql", SourceDir: "public", File: "public/001.sql", MigrationType: types.MigrationTypeSingleMigration}
	db1 := types.MigrationDB{Migration: m1, Schema: "public", AppliedAt: time.Now()}

	m2 := types.Migration{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantMigration}
	db2 := types.MigrationDB{Migration: m2, Schema: "abc", AppliedAt: time.Now()}

	db3 := types.MigrationDB{Migration: m2, Schema: "def", AppliedAt: time.Now()}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleMigration}
	db4 := types.MigrationDB{Migration: m4, Schema: "ref", AppliedAt: time.Now()}

	dbs := []types.MigrationDB{db1, db2, db3, db4}

	coordinator := &coordinator{
		connector: newMockedConnector(context.TODO(), nil),
		loader:    newMockedDiskLoader(context.TODO(), nil),
		notifier:  newMockedNotifier(context.TODO(), nil),
	}
	migrations := coordinator.flattenAppliedMigrations(dbs)

	assert.Equal(t, []types.Migration{m1, m2, m4}, migrations)
}

func TestMigrationsFlattenMigrationDBs2(t *testing.T) {
	m2 := types.Migration{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantMigration}
	db2 := types.MigrationDB{Migration: m2, Schema: "abc", AppliedAt: time.Now()}

	db3 := types.MigrationDB{Migration: m2, Schema: "def", AppliedAt: time.Now()}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleMigration}
	db4 := types.MigrationDB{Migration: m4, Schema: "ref", AppliedAt: time.Now()}

	dbs := []types.MigrationDB{db2, db3, db4}

	coordinator := &coordinator{}
	migrations := coordinator.flattenAppliedMigrations(dbs)

	assert.Equal(t, []types.Migration{m2, m4}, migrations)
}

func TestMigrationsFlattenMigrationDBs3(t *testing.T) {
	m1 := types.Migration{Name: "001.sql", SourceDir: "public", File: "public/001.sql", MigrationType: types.MigrationTypeSingleMigration}
	db1 := types.MigrationDB{Migration: m1, Schema: "public", AppliedAt: time.Now()}

	m2 := types.Migration{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantMigration}
	db2 := types.MigrationDB{Migration: m2, Schema: "abc", AppliedAt: time.Now()}

	db3 := types.MigrationDB{Migration: m2, Schema: "def", AppliedAt: time.Now()}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleMigration}
	db4 := types.MigrationDB{Migration: m4, Schema: "ref", AppliedAt: time.Now()}

	m5 := types.Migration{Name: "global-stored-procedure1.sql", SourceDir: "public", File: "public-scripts/global-stored-procedure1.sql", MigrationType: types.MigrationTypeSingleScript}
	db5 := types.MigrationDB{Migration: m5, Schema: "public", AppliedAt: time.Now()}

	m6 := types.Migration{Name: "global-stored-procedure2.sql", SourceDir: "public", File: "public-scripts/global-stored-procedure2sql", MigrationType: types.MigrationTypeSingleScript}
	db6 := types.MigrationDB{Migration: m6, Schema: "public", AppliedAt: time.Now()}

	m7 := types.Migration{Name: "002.sql", SourceDir: "tenants-scripts", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantMigration}
	db7 := types.MigrationDB{Migration: m7, Schema: "abc", AppliedAt: time.Now()}

	db8 := types.MigrationDB{Migration: m7, Schema: "def", AppliedAt: time.Now()}

	dbs := []types.MigrationDB{db1, db2, db3, db4, db5, db6, db7, db8}

	coordinator := &coordinator{}
	migrations := coordinator.flattenAppliedMigrations(dbs)

	assert.Equal(t, []types.Migration{m1, m2, m4, m5, m6, m7}, migrations)
}

func TestComputeMigrationsToApply(t *testing.T) {
	mdef1 := types.Migration{Name: "a", SourceDir: "a", File: "a", MigrationType: types.MigrationTypeSingleMigration}
	mdef2 := types.Migration{Name: "b", SourceDir: "b", File: "b", MigrationType: types.MigrationTypeTenantMigration}
	mdef3 := types.Migration{Name: "c", SourceDir: "c", File: "c", MigrationType: types.MigrationTypeTenantMigration}
	mdef4 := types.Migration{Name: "d", SourceDir: "d", File: "d", MigrationType: types.MigrationTypeSingleMigration}
	mdef5 := types.Migration{Name: "e", SourceDir: "e", File: "e", MigrationType: types.MigrationTypeSingleScript}
	mdef6 := types.Migration{Name: "f", SourceDir: "f", File: "f", MigrationType: types.MigrationTypeSingleScript}
	mdef7 := types.Migration{Name: "g", SourceDir: "g", File: "g", MigrationType: types.MigrationTypeTenantScript}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, mdef4, mdef5, mdef6, mdef7}
	dbMigrations := []types.MigrationDB{{Migration: mdef1, Schema: "a", AppliedAt: time.Now()}, {Migration: mdef2, Schema: "abc", AppliedAt: time.Now()}, {Migration: mdef2, Schema: "def", AppliedAt: time.Now()}, {Migration: mdef5, Schema: "e", AppliedAt: time.Now()}, {Migration: mdef6, Schema: "f", AppliedAt: time.Now()}, {Migration: mdef7, Schema: "abc", AppliedAt: time.Now()}, {Migration: mdef7, Schema: "def", AppliedAt: time.Now()}}

	coordinator := &coordinator{
		ctx:       context.TODO(),
		connector: newMockedConnector(context.TODO(), nil),
		loader:    newMockedDiskLoader(context.TODO(), nil),
		notifier:  newMockedNotifier(context.TODO(), nil),
	}
	migrations := coordinator.computeMigrationsToApply(diskMigrations, dbMigrations)

	// that should be 5 now...
	assert.Len(t, migrations, 5)

	// 2 migrations
	assert.Equal(t, "c", migrations[0].File)
	assert.Equal(t, "d", migrations[1].File)
	// 3 scripts
	assert.Equal(t, "e", migrations[2].File)
	assert.Equal(t, "f", migrations[3].File)
	assert.Equal(t, "g", migrations[4].File)
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

	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantMigration}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleMigration}
	mdef3 := types.Migration{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleMigration}

	dev1 := types.Migration{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantMigration}
	dev1p1 := types.Migration{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleMigration}
	dev1p2 := types.Migration{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleMigration}

	dev2 := types.Migration{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantMigration}
	dev2p := types.Migration{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleMigration}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, dev1, dev1p1, dev1p2, dev2, dev2p}
	dbMigrations := []types.MigrationDB{{Migration: mdef1, Schema: "abc", AppliedAt: time.Now()}, {Migration: mdef1, Schema: "def", AppliedAt: time.Now()}, {Migration: mdef2, Schema: "public", AppliedAt: time.Now()}, {Migration: mdef3, Schema: "public", AppliedAt: time.Now()}, {Migration: dev2, Schema: "abc", AppliedAt: time.Now()}, {Migration: dev2, Schema: "def", AppliedAt: time.Now()}, {Migration: dev2p, Schema: "public", AppliedAt: time.Now()}}

	coordinator := &coordinator{
		ctx:       context.TODO(),
		connector: newMockedConnector(context.TODO(), nil),
		loader:    newMockedDiskLoader(context.TODO(), nil),
		notifier:  newMockedNotifier(context.TODO(), nil),
	}
	migrations := coordinator.computeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 3)

	assert.Equal(t, dev1.File, migrations[0].File)
	assert.Equal(t, dev1p1.File, migrations[1].File)
	assert.Equal(t, dev1p2.File, migrations[2].File)
}

func TestFilterTenantMigrations(t *testing.T) {
	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantMigration}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleMigration}
	mdef3 := types.Migration{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleMigration}

	dev1 := types.Migration{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantMigration}
	dev1p1 := types.Migration{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleMigration}
	dev1p2 := types.Migration{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleMigration}

	dev2 := types.Migration{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantMigration}
	dev2p := types.Migration{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleMigration}

	script := types.Migration{Name: "20181120", SourceDir: "tenants-script", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantScript}
	scriptp := types.Migration{Name: "20181120", SourceDir: "public-script", File: "public/20181120", MigrationType: types.MigrationTypeSingleScript}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, dev1, dev1p1, dev1p2, dev2, dev2p, script, scriptp}

	coordinator := &coordinator{
		ctx:       context.TODO(),
		connector: newMockedConnector(context.TODO(), nil),
		loader:    newMockedDiskLoader(context.TODO(), nil),
		notifier:  newMockedNotifier(context.TODO(), nil),
	}
	migrations := coordinator.filterTenantMigrations(diskMigrations)

	assert.Len(t, migrations, 4)

	assert.Equal(t, mdef1.File, migrations[0].File)
	assert.Equal(t, types.MigrationTypeTenantMigration, migrations[0].MigrationType)
	assert.Equal(t, dev1.File, migrations[1].File)
	assert.Equal(t, types.MigrationTypeTenantMigration, migrations[1].MigrationType)
	assert.Equal(t, dev2.File, migrations[2].File)
	assert.Equal(t, types.MigrationTypeTenantMigration, migrations[2].MigrationType)
	assert.Equal(t, script.File, migrations[3].File)
	assert.Equal(t, types.MigrationTypeTenantScript, migrations[3].MigrationType)
}

func TestIntersect(t *testing.T) {
	mdef1 := types.Migration{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantMigration}
	mdef2 := types.Migration{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleMigration}
	mdef3 := types.Migration{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleMigration}

	dev1 := types.Migration{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantMigration}
	dev1p1 := types.Migration{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleMigration}
	dev1p2 := types.Migration{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleMigration}

	dev2 := types.Migration{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantMigration}
	dev2p := types.Migration{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleMigration}

	diskMigrations := []types.Migration{mdef1, mdef2, mdef3, dev1, dev1p1, dev1p2, dev2, dev2p}
	dbMigrations := []types.Migration{mdef1, mdef2, mdef3, dev2, dev2p}

	coordinator := &coordinator{}
	intersect := coordinator.intersect(diskMigrations, dbMigrations)
	assert.Len(t, intersect, 5)
	for i := range intersect {
		assert.Equal(t, intersect[i].source, intersect[i].applied)
		assert.Equal(t, intersect[i].source, dbMigrations[i])
	}
}

func TestVerifySourceMigrationsCheckSumsOK(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	verified, offendingMigrations := coordinator.VerifySourceMigrationsCheckSums()
	assert.True(t, verified)
	assert.Empty(t, offendingMigrations)
}

func TestVerifySourceMigrationsCheckSumsKO(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newBrokenCheckSumMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	verified, offendingMigrations := coordinator.VerifySourceMigrationsCheckSums()
	assert.False(t, verified)
	assert.Equal(t, coordinator.GetSourceMigrations()[0], offendingMigrations[0])
}

func TestApplyMigrations(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	_, appliedMigrations := coordinator.ApplyMigrations()
	assert.Len(t, appliedMigrations, 1)
	assert.Equal(t, coordinator.GetSourceMigrations()[1], appliedMigrations[0])
}

func TestAddTenantAndApplyMigrations(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	_, appliedMigrations := coordinator.AddTenantAndApplyMigrations("new")
	assert.Len(t, appliedMigrations, 1)
	assert.Equal(t, coordinator.GetSourceMigrations()[1], appliedMigrations[0])
}

func TestGetTenants(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	tenants := coordinator.GetTenants()
	assert.Equal(t, []string{"a", "b", "c"}, tenants)
}
