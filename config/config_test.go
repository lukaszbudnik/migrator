package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func noopLogger() *log.Logger {
	log := log.New(ioutil.Discard, "", 0)
	return log
}

func TestFromFile(t *testing.T) {
	config, err := FromFile("../test/migrator-test.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "test/migrations", config.BaseDir)
	assert.Equal(t, "select name from migrator.migrator_tenants", config.TenantSelectSQL)
	assert.Equal(t, "postgres", config.Driver)
	assert.Equal(t, "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable", config.DataSource)
	assert.Equal(t, []string{"tenants"}, config.TenantSchemas)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleSchemas)
	assert.Equal(t, "8811", config.Port)
	assert.Equal(t, "{schema}", config.SchemaPlaceHolder)
}

func TestWithEnvFromFile(t *testing.T) {
	config, err := FromFile("../test/migrator-test-envs.yaml")
	assert.Nil(t, err)
	assert.Equal(t, os.Getenv("HOME"), config.BaseDir)
	assert.Equal(t, os.Getenv("PATH"), config.TenantSelectSQL)
	assert.Equal(t, os.Getenv("GOPATH"), config.TenantInsertSQL)
	assert.Equal(t, os.Getenv("PWD"), config.Driver)
	assert.Equal(t, os.Getenv("TERM"), config.DataSource)
	assert.Equal(t, os.Getenv("_"), config.Port)
	assert.Equal(t, os.Getenv("SHLVL"), config.SlackWebHook)
	assert.Equal(t, os.Getenv("USER"), config.SchemaPlaceHolder)
	assert.Equal(t, []string{"tenants"}, config.TenantSchemas)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleSchemas)
}

func TestConfigString(t *testing.T) {
	config := &Config{"/opt/app/migrations", "postgres", "user=p dbname=db host=localhost", "select abc", "insert into table", ":tenant", []string{"ref"}, []string{"tenants"}, "8181", "https://hooks.slack.com/services/TTT/BBB/XXX"}
	// check if go naming convention applies
	expected := `baseDir: /opt/app/migrations
driver: postgres
dataSource: user=p dbname=db host=localhost
tenantSelectSQL: select abc
tenantInsertSQL: insert into table
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
	config, err := FromFile("../test/empty.yaml")
	assert.Nil(t, config)
	assert.IsType(t, (validator.ErrorMap)(nil), err, "Should error because of validation errors")
}

func TestConfigPanicFromNonExistingFile(t *testing.T) {
	config, err := FromFile("abcxyz.yaml")
	assert.Nil(t, config)
	assert.IsType(t, (*os.PathError)(nil), err, "Should error because non-existing file")
}

func TestConfigFromWrongSyntaxFile(t *testing.T) {
	config, err := FromFile("../README.md")
	assert.Nil(t, config)
	assert.IsType(t, (*yaml.TypeError)(nil), err, "Should panic because of wrong yaml syntax")
}
