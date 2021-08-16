package coordinator

import (
	"context"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/assert"

	"github.com/lukaszbudnik/migrator/types"
)

func TestMigrationsFlattenMigrationDBs1(t *testing.T) {
	m1 := types.Migration{Name: "001.sql", SourceDir: "public", File: "public/001.sql", MigrationType: types.MigrationTypeSingleMigration}
	db1 := types.DBMigration{Migration: m1, Schema: "public", AppliedAt: graphql.Time{Time: time.Now()}}

	m2 := types.Migration{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantMigration}
	db2 := types.DBMigration{Migration: m2, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}

	db3 := types.DBMigration{Migration: m2, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleMigration}
	db4 := types.DBMigration{Migration: m4, Schema: "ref", AppliedAt: graphql.Time{Time: time.Now()}}

	dbs := []types.DBMigration{db1, db2, db3, db4}

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
	db2 := types.DBMigration{Migration: m2, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}

	db3 := types.DBMigration{Migration: m2, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleMigration}
	db4 := types.DBMigration{Migration: m4, Schema: "ref", AppliedAt: graphql.Time{Time: time.Now()}}

	dbs := []types.DBMigration{db2, db3, db4}

	coordinator := &coordinator{}
	migrations := coordinator.flattenAppliedMigrations(dbs)

	assert.Equal(t, []types.Migration{m2, m4}, migrations)
}

func TestMigrationsFlattenMigrationDBs3(t *testing.T) {
	m1 := types.Migration{Name: "001.sql", SourceDir: "public", File: "public/001.sql", MigrationType: types.MigrationTypeSingleMigration}
	db1 := types.DBMigration{Migration: m1, Schema: "public", AppliedAt: graphql.Time{Time: time.Now()}}

	m2 := types.Migration{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantMigration}
	db2 := types.DBMigration{Migration: m2, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}

	db3 := types.DBMigration{Migration: m2, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}

	m4 := types.Migration{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleMigration}
	db4 := types.DBMigration{Migration: m4, Schema: "ref", AppliedAt: graphql.Time{Time: time.Now()}}

	m5 := types.Migration{Name: "global-stored-procedure1.sql", SourceDir: "public", File: "public-scripts/global-stored-procedure1.sql", MigrationType: types.MigrationTypeSingleScript}
	db5 := types.DBMigration{Migration: m5, Schema: "public", AppliedAt: graphql.Time{Time: time.Now()}}

	m6 := types.Migration{Name: "global-stored-procedure2.sql", SourceDir: "public", File: "public-scripts/global-stored-procedure2sql", MigrationType: types.MigrationTypeSingleScript}
	db6 := types.DBMigration{Migration: m6, Schema: "public", AppliedAt: graphql.Time{Time: time.Now()}}

	m7 := types.Migration{Name: "002.sql", SourceDir: "tenants-scripts", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantMigration}
	db7 := types.DBMigration{Migration: m7, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}

	db8 := types.DBMigration{Migration: m7, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}

	dbs := []types.DBMigration{db1, db2, db3, db4, db5, db6, db7, db8}

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
	dbMigrations := []types.DBMigration{{Migration: mdef1, Schema: "a", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef2, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef2, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef5, Schema: "e", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef6, Schema: "f", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef7, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef7, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}}

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
	// development done in parallel, 2 devs fork from main
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
	dbMigrations := []types.DBMigration{{Migration: mdef1, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef1, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef2, Schema: "public", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: mdef3, Schema: "public", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: dev2, Schema: "abc", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: dev2, Schema: "def", AppliedAt: graphql.Time{Time: time.Now()}}, {Migration: dev2p, Schema: "public", AppliedAt: graphql.Time{Time: time.Now()}}}

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
	assert.Equal(t, coordinator.GetSourceMigrations(nil)[0], offendingMigrations[0])
}

func TestVerifySourceMigrationsAndScriptsCheckSumsOK(t *testing.T) {
	coordinator := New(context.TODO(), nil, newDifferentScriptCheckSumMockedConnector, newDifferentScriptCheckSumMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	verified, offendingMigrations := coordinator.VerifySourceMigrationsCheckSums()
	assert.True(t, verified)
	assert.Empty(t, offendingMigrations)
}

func TestGetTenants(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	tenants := coordinator.GetTenants()
	a := types.Tenant{Name: "a"}
	b := types.Tenant{Name: "b"}
	c := types.Tenant{Name: "c"}
	assert.Equal(t, []types.Tenant{a, b, c}, tenants)
}

func TestGetVersions(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	versions := coordinator.GetVersions()

	assert.Equal(t, int32(12), versions[0].ID)
	assert.Equal(t, int32(121), versions[1].ID)
	assert.Equal(t, int32(122), versions[2].ID)
}

func TestGetVersionByID(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	version, _ := coordinator.GetVersionByID(123)

	assert.Equal(t, int32(123), version.ID)
}

func TestGetVersionsByFile(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	versions := coordinator.GetVersionsByFile("tenants/abc.sql")

	assert.Equal(t, int32(12), versions[0].ID)
}

func TestGetMigrationByID(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newMockedNotifier)
	defer coordinator.Dispose()
	migration, _ := coordinator.GetDBMigrationByID(456)

	assert.Equal(t, int32(456), migration.ID)
}

func TestGetSourceMigrationByFile(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	file := "source/201602220001.sql"
	migration, err := coordinator.GetSourceMigrationByFile(file)
	assert.Nil(t, err)
	assert.Equal(t, file, migration.File)
}

func TestGetSourceMigrationByFileNotFound(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	file := "xyz/201602220001.sql"
	_, err := coordinator.GetSourceMigrationByFile(file)
	assert.NotNil(t, err)
	assert.Equal(t, "Source migration not found: xyz/201602220001.sql", err.Error())
}

func TestGetSourceMigrationsFilterMigrationType(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	migrationType := types.MigrationTypeSingleMigration
	filters := SourceMigrationFilters{
		MigrationType: &migrationType,
	}
	migrations := coordinator.GetSourceMigrations(&filters)
	assert.True(t, len(migrations) == 4)
}

func TestGetSourceMigrationsFilterMigrationTypeSourceDir(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	migrationType := types.MigrationTypeSingleMigration
	sourceDir := "source"
	filters := SourceMigrationFilters{
		MigrationType: &migrationType,
		SourceDir:     &sourceDir,
	}
	migrations := coordinator.GetSourceMigrations(&filters)
	assert.True(t, len(migrations) == 3)
}

func TestGetSourceMigrationsFilterMigrationTypeName(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	migrationType := types.MigrationTypeSingleMigration
	name := "201602220001.sql"
	filters := SourceMigrationFilters{
		MigrationType: &migrationType,
		Name:          &name,
	}
	migrations := coordinator.GetSourceMigrations(&filters)
	assert.True(t, len(migrations) == 2)
}

func TestGetSourceMigrationsFilterFile(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	file := "source/201602220001.sql"
	filters := SourceMigrationFilters{
		File: &file,
	}
	migrations := coordinator.GetSourceMigrations(&filters)
	assert.True(t, len(migrations) == 1)
}

func TestCreateVersion(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	results := coordinator.CreateVersion("commit-sha", types.ActionApply, false)
	assert.NotNil(t, results)
	assert.NotNil(t, results.Summary)
	assert.NotNil(t, results.Version)
}

func TestCreateTenant(t *testing.T) {
	coordinator := New(context.TODO(), nil, newMockedConnector, newMockedDiskLoader, newErrorMockedNotifier)
	defer coordinator.Dispose()
	results := coordinator.CreateTenant("commit-sha", types.ActionSync, true, "NewTenant")
	assert.NotNil(t, results)
	assert.NotNil(t, results.Summary)
	assert.NotNil(t, results.Version)
}
