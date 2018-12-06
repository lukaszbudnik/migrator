package migrations

import (
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMigrationsFlattenMigrationDBs1(t *testing.T) {
	m1 := types.MigrationDefinition{Name: "001.sql", SourceDir: "public", File: "public/001.sql", MigrationType: types.MigrationTypeSingleSchema}
	db1 := types.MigrationDB{MigrationDefinition: m1, Schema: "public", Created: time.Now()}

	m2 := types.MigrationDefinition{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantSchema}
	db2 := types.MigrationDB{MigrationDefinition: m2, Schema: "abc", Created: time.Now()}

	db3 := types.MigrationDB{MigrationDefinition: m2, Schema: "def", Created: time.Now()}

	m4 := types.MigrationDefinition{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleSchema}
	db4 := types.MigrationDB{MigrationDefinition: m4, Schema: "ref", Created: time.Now()}

	dbs := []types.MigrationDB{db1, db2, db3, db4}

	migrations := flattenMigrationDBs(dbs)

	assert.Equal(t, []types.MigrationDefinition{m1, m2, m4}, migrations)

}

func TestMigrationsFlattenMigrationDBs2(t *testing.T) {

	m2 := types.MigrationDefinition{Name: "002.sql", SourceDir: "tenants", File: "tenants/002.sql", MigrationType: types.MigrationTypeTenantSchema}
	db2 := types.MigrationDB{MigrationDefinition: m2, Schema: "abc", Created: time.Now()}

	db3 := types.MigrationDB{MigrationDefinition: m2, Schema: "def", Created: time.Now()}

	m4 := types.MigrationDefinition{Name: "003.sql", SourceDir: "ref", File: "ref/003.sql", MigrationType: types.MigrationTypeSingleSchema}
	db4 := types.MigrationDB{MigrationDefinition: m4, Schema: "ref", Created: time.Now()}

	dbs := []types.MigrationDB{db2, db3, db4}

	migrations := flattenMigrationDBs(dbs)

	assert.Equal(t, []types.MigrationDefinition{m2, m4}, migrations)

}

func TestComputeMigrationsToApply(t *testing.T) {
	mdef1 := types.MigrationDefinition{Name: "a", SourceDir: "a", File: "a", MigrationType: types.MigrationTypeSingleSchema}
	mdef2 := types.MigrationDefinition{Name: "b", SourceDir: "b", File: "b", MigrationType: types.MigrationTypeTenantSchema}
	mdef3 := types.MigrationDefinition{Name: "c", SourceDir: "c", File: "c", MigrationType: types.MigrationTypeTenantSchema}
	mdef4 := types.MigrationDefinition{Name: "d", SourceDir: "d", File: "d", MigrationType: types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{{MigrationDefinition: mdef1, Contents: ""}, {MigrationDefinition: mdef2, Contents: ""}, {MigrationDefinition: mdef3, Contents: ""}, {MigrationDefinition: mdef4, Contents: ""}}
	dbMigrations := []types.MigrationDB{{MigrationDefinition: mdef1, Schema: "a", Created: time.Now()}, {MigrationDefinition: mdef2, Schema: "abc", Created: time.Now()}, {MigrationDefinition: mdef2, Schema: "def", Created: time.Now()}}
	migrations := ComputeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 2)

	assert.Equal(t, "c", migrations[0].File)
	assert.Equal(t, "d", migrations[1].File)
}

func TestFilterTenantMigrations(t *testing.T) {
	mdef1 := types.MigrationDefinition{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantSchema}
	mdef2 := types.MigrationDefinition{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleSchema}
	mdef3 := types.MigrationDefinition{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleSchema}

	dev1 := types.MigrationDefinition{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantSchema}
	dev1p1 := types.MigrationDefinition{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleSchema}
	dev1p2 := types.MigrationDefinition{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleSchema}

	dev2 := types.MigrationDefinition{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantSchema}
	dev2p := types.MigrationDefinition{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{{MigrationDefinition: mdef1, Contents: ""}, {MigrationDefinition: mdef2, Contents: ""}, {MigrationDefinition: mdef3, Contents: ""}, {MigrationDefinition: dev1, Contents: ""}, {MigrationDefinition: dev1p1, Contents: ""}, {MigrationDefinition: dev1p2, Contents: ""}, {MigrationDefinition: dev2, Contents: ""}, {MigrationDefinition: dev2p, Contents: ""}}
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

	mdef1 := types.MigrationDefinition{Name: "20181111", SourceDir: "tenants", File: "tenants/20181111", MigrationType: types.MigrationTypeTenantSchema}
	mdef2 := types.MigrationDefinition{Name: "20181111", SourceDir: "public", File: "public/20181111", MigrationType: types.MigrationTypeSingleSchema}
	mdef3 := types.MigrationDefinition{Name: "20181112", SourceDir: "public", File: "public/20181112", MigrationType: types.MigrationTypeSingleSchema}

	dev1 := types.MigrationDefinition{Name: "20181119", SourceDir: "tenants", File: "tenants/20181119", MigrationType: types.MigrationTypeTenantSchema}
	dev1p1 := types.MigrationDefinition{Name: "201811190", SourceDir: "public", File: "public/201811190", MigrationType: types.MigrationTypeSingleSchema}
	dev1p2 := types.MigrationDefinition{Name: "20181191", SourceDir: "public", File: "public/201811191", MigrationType: types.MigrationTypeSingleSchema}

	dev2 := types.MigrationDefinition{Name: "20181120", SourceDir: "tenants", File: "tenants/20181120", MigrationType: types.MigrationTypeTenantSchema}
	dev2p := types.MigrationDefinition{Name: "20181120", SourceDir: "public", File: "public/20181120", MigrationType: types.MigrationTypeSingleSchema}

	diskMigrations := []types.Migration{{MigrationDefinition: mdef1, Contents: ""}, {MigrationDefinition: mdef2, Contents: ""}, {MigrationDefinition: mdef3, Contents: ""}, {MigrationDefinition: dev1, Contents: ""}, {MigrationDefinition: dev1p1, Contents: ""}, {MigrationDefinition: dev1p2, Contents: ""}, {MigrationDefinition: dev2, Contents: ""}, {MigrationDefinition: dev2p, Contents: ""}}
	dbMigrations := []types.MigrationDB{{MigrationDefinition: mdef1, Schema: "abc", Created: time.Now()}, {MigrationDefinition: mdef1, Schema: "def", Created: time.Now()}, {MigrationDefinition: mdef2, Schema: "public", Created: time.Now()}, {MigrationDefinition: mdef3, Schema: "public", Created: time.Now()}, {MigrationDefinition: dev2, Schema: "abc", Created: time.Now()}, {MigrationDefinition: dev2, Schema: "def", Created: time.Now()}, {MigrationDefinition: dev2p, Schema: "public", Created: time.Now()}}
	migrations := ComputeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 3)

	assert.Equal(t, dev1.File, migrations[0].File)
	assert.Equal(t, dev1p1.File, migrations[1].File)
	assert.Equal(t, dev1p2.File, migrations[2].File)
}
