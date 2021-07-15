package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

func TestFromFile(t *testing.T) {
	config, err := FromFile("../test/migrator-test.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "test/migrations", config.BaseLocation)
	assert.Equal(t, "select name from migrator.migrator_tenants", config.TenantSelectSQL)
	assert.Equal(t, "postgres", config.Driver)
	assert.Equal(t, "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable", config.DataSource)
	assert.Equal(t, []string{"tenants"}, config.TenantMigrations)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleMigrations)
	assert.Equal(t, "8811", config.Port)
	assert.Equal(t, "{schema}", config.SchemaPlaceHolder)
	assert.Equal(t, "https://slack.com/api/api.test", config.WebHookURL)
	assert.Equal(t, []string{"Authorization: Basic QWxhZGRpbjpPcGVuU2VzYW1l", "Content-Type: application/json", "X-CustomHeader: value1,value2"}, config.WebHookHeaders)
}

func TestWithEnvFromFile(t *testing.T) {
	os.Setenv("COMMIT_SHA", "62fd74506651982fe317721d7e07145f8c2fa166")
	config, err := FromFile("../test/migrator-test-envs.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "s3://bucket-name/application-x/"+os.Getenv("TERM")+"/"+os.Getenv("COMMIT_SHA"), config.BaseLocation)
	assert.Equal(t, os.Getenv("PATH"), config.TenantSelectSQL)
	assert.Equal(t, os.Getenv("GOPATH"), config.TenantInsertSQL)
	assert.Equal(t, os.Getenv("PWD"), config.Driver)
	assert.Equal(t, fmt.Sprintf("lets_assume_password=%v&and_something_else=%v&param=value", os.Getenv("HOME"), os.Getenv("USER")), config.DataSource)
	assert.Equal(t, os.Getenv("_"), config.Port)
	assert.Equal(t, os.Getenv("USER"), config.SchemaPlaceHolder)
	assert.Equal(t, []string{"tenants"}, config.TenantMigrations)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleMigrations)
	assert.Equal(t, os.Getenv("SHLVL"), config.WebHookURL)
	assert.Equal(t, fmt.Sprintf("X-Security-Token: %v", os.Getenv("USER")), config.WebHookHeaders[0])
}

func TestConfigString(t *testing.T) {
	config := &Config{"/opt/app/migrations", "postgres", "user=p dbname=db host=localhost", "select abc", "insert into table", ":tenant", []string{"ref"}, []string{"tenants"}, []string{"procedures"}, []string{}, "8181", "", "https://hooks.slack.com/services/TTT/BBB/XXX", []string{}, `{"text": "Results are: ${summary}"}`}
	// check if go naming convention applies
	expected := `baseLocation: /opt/app/migrations
driver: postgres
dataSource: user=p dbname=db host=localhost
tenantSelectSQL: select abc
tenantInsertSQL: insert into table
schemaPlaceHolder: :tenant
singleMigrations:
- ref
tenantMigrations:
- tenants
singleScripts:
- procedures
port: "8181"
webHookURL: https://hooks.slack.com/services/TTT/BBB/XXX
webHookTemplate: '{"text": "Results are: ${summary}"}'`
	actual := fmt.Sprintf("%v", config)
	assert.Equal(t, expected, actual)
}

func TestConfigReadFromEmptyFileError(t *testing.T) {
	config, err := FromFile("../test/empty.yaml")
	assert.Nil(t, config)
	assert.IsType(t, (validator.ValidationErrors)(nil), err, "Should error because of validation errors")
}

func TestConfigReadFromNonExistingFileError(t *testing.T) {
	config, err := FromFile("abcxyz.yaml")
	assert.Nil(t, config)
	assert.IsType(t, (*os.PathError)(nil), err, "Should error because non-existing file")
}

func TestConfigFromWrongSyntaxFile(t *testing.T) {
	config, err := FromFile("../Dockerfile")
	assert.Nil(t, config)
	assert.IsType(t, (*yaml.TypeError)(nil), err, "Should error because of wrong yaml syntax")
}
