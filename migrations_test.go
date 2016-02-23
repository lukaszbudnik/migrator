package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMigrationsFlattenDBMigrations1(t *testing.T) {

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

func TestMigrationsFlattenDBMigrations2(t *testing.T) {

	m2 := MigrationDefinition{"002.sql", "tenants", "tenants/002.sql", ModeTenantSchema}
	db2 := DBMigration{m2, "abc", time.Now()}

	db3 := DBMigration{m2, "def", time.Now()}

	m4 := MigrationDefinition{"003.sql", "ref", "ref/003.sql", ModeSingleSchema}
	db4 := DBMigration{m4, "ref", time.Now()}

	dbs := []DBMigration{db2, db3, db4}

	migrations := flattenDBMigrations(dbs)

	assert.Equal(t, []MigrationDefinition{m2, m4}, migrations)

}

func TestComputeMigrationsToApply(t *testing.T) {
	mdef1 := MigrationDefinition{"a", "a", "a", ModeSingleSchema}
	mdef2 := MigrationDefinition{"b", "b", "b", ModeTenantSchema}
	mdef3 := MigrationDefinition{"c", "c", "c", ModeTenantSchema}
	mdef4 := MigrationDefinition{"d", "d", "d", ModeSingleSchema}

	diskMigrations := []Migration{{mdef1, ""}, {mdef2, ""}, {mdef3, ""}, {mdef4, ""}}
	dbMigrations := []DBMigration{{mdef1, "a", time.Now()}, {mdef2, "abc", time.Now()}, {mdef2, "def", time.Now()}}
	migrations := computeMigrationsToApply(diskMigrations, dbMigrations)

	assert.Len(t, migrations, 2)

	assert.Equal(t, "c", migrations[0].File)
	assert.Equal(t, "d", migrations[1].File)
}
