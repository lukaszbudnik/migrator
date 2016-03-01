package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTypesDBTenantsString(t *testing.T) {
	dbTenants := []string{"abcabc", "dedededededededededede", "opopopop"}
	expected := `+------------------------------+
| abcabc                       |
| dedededededededededede       |
| opopopop                     |
+------------------------------+`

	actual := DBTenantsString(dbTenants)

	assert.Equal(t, expected, actual)
}

func TestTypesMigrationsString(t *testing.T) {

	m1 := MigrationDefinition{"201602220000.sql", "source", "source/201602220000.sql", MigrationTypeSingleSchema}
	m2 := MigrationDefinition{"201602220001.sql", "tenants", "tenants/201602220001.sql", MigrationTypeTenantSchema}
	m3 := MigrationDefinition{"201602220002.sql", "tenants", "tenants/201602220002.sql", MigrationTypeTenantSchema}
	var ms = []Migration{{m1, ""}, {m2, ""}, {m3, ""}}

	expected := `+---------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Type |
+---------------------------------------------------------------------------+
| source     | 201602220000.sql     | source/201602220000.sql        |    1 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       |    2 |
| tenants    | 201602220002.sql     | tenants/201602220002.sql       |    2 |
+---------------------------------------------------------------------------+`
	actual := MigrationsString(ms)

	assert.Equal(t, expected, actual)
}

func TestTypesMigrationsEmptyArrayString(t *testing.T) {

	var ms = []Migration{}

	expected := `+---------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Type |
+---------------------------------------------------------------------------+
| Empty                                                                     |
+---------------------------------------------------------------------------+`
	actual := MigrationsString(ms)

	assert.Equal(t, expected, actual)
}

func TestTypesMigrationDBsString(t *testing.T) {
	m1 := MigrationDefinition{"201602220000.sql", "source", "source/201602220000.sql", MigrationTypeSingleSchema}
	m2 := MigrationDefinition{"201602220001.sql", "tenants", "tenants/201602220001.sql", MigrationTypeTenantSchema}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	d2 := time.Date(2016, 02, 22, 16, 41, 2, 456, time.UTC)
	var ms = []MigrationDB{{m1, "source", d1}, {m2, "abc", d2}, {m2, "def", d2}}

	expected := `+---------------------------------------------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Schema     | Created              | Type |
+---------------------------------------------------------------------------------------------------------------+
| source     | 201602220000.sql     | source/201602220000.sql        | source     | 2016-02-22 16:41:01  |    1 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       | abc        | 2016-02-22 16:41:02  |    2 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       | def        | 2016-02-22 16:41:02  |    2 |
+---------------------------------------------------------------------------------------------------------------+`
	actual := MigrationDBsString(ms)

	assert.Equal(t, expected, actual)
}

func TestTypesMigrationDBsEmptyArrayString(t *testing.T) {
	var ms = []MigrationDB{}

	expected := `+---------------------------------------------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Schema     | Created              | Type |
+---------------------------------------------------------------------------------------------------------------+
| Empty                                                                                                         |
+---------------------------------------------------------------------------------------------------------------+`
	actual := MigrationDBsString(ms)

	assert.Equal(t, expected, actual)
}
