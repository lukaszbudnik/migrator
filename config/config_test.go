package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFromFile(t *testing.T) {
	config := FromFile("../test/migrator-test.yaml")
	assert.Equal(t, "test/migrations", config.BaseDir)
	assert.Equal(t, "select name from public.migrator_tenants", config.TenantsSQL)
	assert.Equal(t, "postgres", config.Driver)
	assert.Equal(t, "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable", config.DataSource)
	assert.Equal(t, []string{"tenants"}, config.TenantSchemas)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleSchemas)
	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "{schema}", config.SchemaPlaceHolder)
}

func TestWithEnvFromFile(t *testing.T) {
	config := FromFile("../test/migrator-test-envs.yaml")
	assert.Equal(t, os.Getenv("HOME"), config.BaseDir)
	assert.Equal(t, os.Getenv("PATH"), config.TenantsSQL)
	assert.Equal(t, os.Getenv("PWD"), config.Driver)
	assert.Equal(t, os.Getenv("TERM"), config.DataSource)
	assert.Equal(t, os.Getenv("_"), config.Port)
	assert.Equal(t, os.Getenv("SHLVL"), config.SlackWebHook)
	assert.Equal(t, os.Getenv("USER"), config.SchemaPlaceHolder)
	assert.Equal(t, []string{"tenants"}, config.TenantSchemas)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleSchemas)
}

func TestConfigString(t *testing.T) {
	config := &Config{"/opt/app/migrations", "postgres", "user=p dbname=db host=localhost", "select abc", ":tenant", []string{"ref"}, []string{"tenants"}, "8181", "https://hooks.slack.com/services/TTT/BBB/XXX"}
	// check if go naming convention applies
	expected := `baseDir: /opt/app/migrations
driver: postgres
dataSource: user=p dbname=db host=localhost
tenantsSql: select abc
schemaPlaceHolder: :tenant
singleSchemas:
- ref
tenantSchemas:
- tenants
port: "8181"
slackWebHook: https://hooks.slack.com/services/TTT/BBB/XXX`
	actual := fmt.Sprintf("%v", config)
	assert.Equal(t, expected, actual)
}

func TestConfigPanicFromEmptyFile(t *testing.T) {
	assert.Panics(t, func() {
		FromFile("../test/empty.yaml")
	}, "Should panic because of validation errors")
}

func TestConfigPanicFromNonExistingFile(t *testing.T) {
	assert.Panics(t, func() {
		FromFile("abcxyz.yaml")
	}, "Should panic because of non-existing file")
}

func TestConfigFromWrongSyntaxFile(t *testing.T) {
	assert.Panics(t, func() {
		FromFile("../README.md")
	}, "Should panic because of wrong yaml syntax")
}
