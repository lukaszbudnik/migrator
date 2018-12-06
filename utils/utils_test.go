package utils

import (
	"fmt"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestContainsFound(t *testing.T) {

	var slice = []string{"abc", "def", "ghi"}

	for _, s := range slice {
		found := Contains(slice, &s)
		assert.True(t, found, fmt.Sprintf("String %#v not found in slice %#v", s, slice))
	}
}

func TestContainsNotFound(t *testing.T) {

	var slice = []string{"abc", "def", "ghi"}
	var other = []string{"", "xyz"}

	for _, s := range other {
		found := Contains(slice, &s)
		assert.False(t, found, fmt.Sprintf("String %#v found in slice %#v", s, slice))
	}
}

func TestTenantArrayToString(t *testing.T) {
	dbTenants := []string{"abcabc", "dedededededededededede", "opopopop"}
	expected := `+------------------------------+
| Name                         |
+------------------------------+
| abcabc                       |
| dedededededededededede       |
| opopopop                     |
+------------------------------+`

	actual := TenantArrayToString(dbTenants)

	assert.Equal(t, expected, actual)
}

func TestMigrationArrayToString(t *testing.T) {

	m1 := types.MigrationDefinition{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema}
	m2 := types.MigrationDefinition{Name: "201602220001.sql", SourceDir: "tenants", File: "tenants/201602220001.sql", MigrationType: types.MigrationTypeTenantSchema}
	m3 := types.MigrationDefinition{Name: "201602220002.sql", SourceDir: "tenants", File: "tenants/201602220002.sql", MigrationType: types.MigrationTypeTenantSchema}
	var ms = []types.Migration{{MigrationDefinition: m1, Contents: ""}, {MigrationDefinition: m2, Contents: ""}, {MigrationDefinition: m3, Contents: ""}}

	expected := `+---------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Type |
+---------------------------------------------------------------------------+
| source     | 201602220000.sql     | source/201602220000.sql        |    1 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       |    2 |
| tenants    | 201602220002.sql     | tenants/201602220002.sql       |    2 |
+---------------------------------------------------------------------------+`
	actual := MigrationArrayToString(ms)

	assert.Equal(t, expected, actual)
}

func TestMigrationArrayToStringEmpty(t *testing.T) {

	var ms = []types.Migration{}

	expected := `+---------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Type |
+---------------------------------------------------------------------------+
+---------------------------------------------------------------------------+`
	actual := MigrationArrayToString(ms)

	assert.Equal(t, expected, actual)
}

func TestMigrationDBArrayToString(t *testing.T) {
	m1 := types.MigrationDefinition{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema}
	m2 := types.MigrationDefinition{Name: "201602220001.sql", SourceDir: "tenants", File: "tenants/201602220001.sql", MigrationType: types.MigrationTypeTenantSchema}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	d2 := time.Date(2016, 02, 22, 16, 41, 2, 456, time.UTC)
	var ms = []types.MigrationDB{{MigrationDefinition: m1, Schema: "source", Created: d1}, {MigrationDefinition: m2, Schema: "abc", Created: d2}, {MigrationDefinition: m2, Schema: "def", Created: d2}}

	expected := `+---------------------------------------------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Schema     | Created              | Type |
+---------------------------------------------------------------------------------------------------------------+
| source     | 201602220000.sql     | source/201602220000.sql        | source     | 2016-02-22 16:41:01  |    1 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       | abc        | 2016-02-22 16:41:02  |    2 |
| tenants    | 201602220001.sql     | tenants/201602220001.sql       | def        | 2016-02-22 16:41:02  |    2 |
+---------------------------------------------------------------------------------------------------------------+`
	actual := MigrationDBArrayToString(ms)

	assert.Equal(t, expected, actual)
}

func TestMigrationDBArrayToStringEmpty(t *testing.T) {
	var ms = []types.MigrationDB{}

	expected := `+---------------------------------------------------------------------------------------------------------------+
| SourceDir  | Name                 | File                           | Schema     | Created              | Type |
+---------------------------------------------------------------------------------------------------------------+
+---------------------------------------------------------------------------------------------------------------+`
	actual := MigrationDBArrayToString(ms)

	assert.Equal(t, expected, actual)
}
