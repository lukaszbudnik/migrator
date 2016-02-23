package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigReadFromFile(t *testing.T) {
	config := readConfigFromFile("test/migrator-test.yaml")
	assert.Equal(t, "test/migrations", config.BaseDir)
	assert.Equal(t, "select name from public.migrator_tenants", config.TenantsSQL)
	assert.Equal(t, "postgres", config.Driver)
	assert.Equal(t, "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable", config.DataSource)
	assert.Equal(t, []string{"tenants"}, config.TenantSchemas)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleSchemas)
}

func TestConfigString(t *testing.T) {
	config := &Config{"/opt/app/migrations", "postgres", "user=p dbname=db host=localhost", "select abc", []string{"ref"}, []string{"tenants"}}
	// check if go naming convention applies
	expected := `baseDir: /opt/app/migrations
driver: postgres
dataSource: user=p dbname=db host=localhost
tenantsSql: select abc
singleSchemas:
- ref
tenantSchemas:
- tenants`
	actual := fmt.Sprintf("%v", config)
	assert.Equal(t, expected, actual)
}

func TestConfigPanicReadFromEmptyFile(t *testing.T) {
	assert.Panics(t, func() {
		readConfigFromFile("test/empty.yaml")
	}, "Should panic because of validation errors")
}

func TestConfigPanicReadFromNonExistingFile(t *testing.T) {
	assert.Panics(t, func() {
		readConfigFromFile("abcxyz.yaml")
	}, "Should panic because of non-existing file")
}

func TestConfigReadFromWrongSyntaxFile(t *testing.T) {
	assert.Panics(t, func() {
		readConfigFromFile("README.md")
	}, "Should panic because of wrong yaml syntax")
}
