package migrations

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"github.com/lukaszbudnik/migrator/types"
)

func TestMigrationsFlattenDBMigrations1(t *testing.T) {

	m1 := types.MigrationDefinition{"001.sql", "public", "public/001.sql", types.ModeSingleSchema}
	db1 := types.DBMigration{m1, "public", time.Now()}

	m2 := types.MigrationDefinition{"002.sql", "tenants", "tenants/002.sql", types.ModeTenantSchema}
	db2 := types.DBMigration{m2, "abc", time.Now()}

	db3 := types.DBMigration{m2, "def", time.Now()}

	m4 := types.MigrationDefinition{"003.sql", "ref", "ref/003.sql", types.ModeSingleSchema}
	db4 := types.DBMigration{m4, "ref", time.Now()}

	dbs := []types.DBMigration{db1, db2, db3, db4}

	migrations := flattenDBMigrations(dbs)

	assert.Equal(t, []types.MigrationDefinition{m1, m2, m4}, migrations)

}

func TestMigrationsFlattenDBMigrations2(t *testing.T) {

	m2 := types.MigrationDefinition{"002.sql", "tenants", "tenants/002.sql", types.ModeTenantSchema}
	db2 := types.DBMigration{m2, "abc", time.Now()}

	db3 := types.DBMigration{m2, "def", time.Now()}

	m4 := types.MigrationDefinition{"003.sql", "ref", "ref/003.sql", types.ModeSingleSchema}
	db4 := types.DBMigration{m4, "ref", time.Now()}

	dbs := []types.DBMigration{db2, db3, db4}

	migrations := flattenDBMigrations(dbs)

	assert.Equal(t, []types.MigrationDefinition{m2, m4}, migrations)

}

func TestComputeMigrationsToApply(t *testing.T) {
	mdef1 := types.MigrationDefinition{"a", "a", "a", types.ModeSingleSchema}
	mdef2 := types.MigrationDefinition{"b", "b", "b", types.ModeTenantSchema}
	mdef3 := types.MigrationDefinition{"c", "c", "c", types.ModeTenantSchema}
	mdef4 := types.MigrationDefinition{"d", "d", "d", types.ModeSingleSchema}

	diskMigrations := []types.Migration{{mdef1, ""}, {mdef2, ""}, {mdef3, ""}, {mdef4, ""}}
	dbMigrations := []types.DBMigration{{mdef1, "a", time.Now()}, {mdef2, "abc", time.Now()}, {mdef2, "def", time.Now()}}
	migrations := ComputeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 2)

	assert.Equal(t, "c", migrations[0].File)
	assert.Equal(t, "d", migrations[1].File)
}
