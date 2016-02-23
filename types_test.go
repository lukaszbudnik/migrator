package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTypesMigrationsString(t *testing.T) {

	m1 := MigrationDefinition{"201602220000.sql", "source", "source/201602220000.sql", ModeSingleSchema}
	m2 := MigrationDefinition{"201602220001.sql", "tenants", "tenants/201602220001.sql", ModeTenantSchema}
	m3 := MigrationDefinition{"201602220002.sql", "tenants", "tenants/201602220002.sql", ModeTenantSchema}
	var ms = []Migration{{m1, ""}, {m2, ""}, {m3, ""}}

	expected := `+---------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Type |
+---------------------------------------------------------------------------+
| source     | 201602220000.sql     | source/201602220000.sql        |    1 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       |    2 |
| tenants    | 201602220002.sql     | tenants/201602220002.sql       |    2 |
+---------------------------------------------------------------------------+`
	actual := migrationsString(ms)

	assert.Equal(t, expected, actual)
}

func TestTypesDBMigrationsString(t *testing.T) {
	m1 := MigrationDefinition{"201602220000.sql", "source", "source/201602220000.sql", ModeSingleSchema}
	m2 := MigrationDefinition{"201602220001.sql", "tenants", "tenants/201602220001.sql", ModeTenantSchema}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	d2 := time.Date(2016, 02, 22, 16, 41, 2, 456, time.UTC)
	var ms = []DBMigration{{m1, "source", d1}, {m2, "abc", d2}, {m2, "def", d2}}

	expected := `+---------------------------------------------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Schema     | Created              | Type |
+---------------------------------------------------------------------------------------------------------------+
| source     | 201602220000.sql     | source/201602220000.sql        | source     | 2016-02-22 16:41:01  |    1 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       | abc        | 2016-02-22 16:41:02  |    2 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       | def        | 2016-02-22 16:41:02  |    2 |
+---------------------------------------------------------------------------------------------------------------+`
	actual := dbMigrationsString(ms)

	assert.Equal(t, expected, actual)
}
