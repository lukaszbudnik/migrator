// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * disk_test.go
// * migrations_test.go

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	unknownAction                    = "unknown"
	nonExistingConfig                = "idontexist"
	configFileWithNonExistingBaseDir = "test/migrator-test-non-existing-base-dir.yaml"
	configFile                       = "test/migrator-test.yaml"
	verbose                          = true
	notVerbose                       = false
)

func TestCliExitUnknownAction(t *testing.T) {
	ret := executeMigrator(&nonExistingConfig, &unknownAction, &notVerbose, CreateConnector, CreateLoader)
	assert.Equal(t, 1, ret)
}

func TestCliPanicReadFromNonExistingConfigFile(t *testing.T) {
	action := applyAction
	assert.Panics(t, func() {
		executeMigrator(&nonExistingConfig, &action, &notVerbose, CreateConnector, CreateLoader)
	}, "Should panic because of non-existing config file")
}

func TestCliPanicReadDiskMigrationsFromNonExistingBaseDir(t *testing.T) {
	action := applyAction
	assert.Panics(t, func() {
		executeMigrator(&configFileWithNonExistingBaseDir, &action, &notVerbose, CreateConnector, CreateLoader)
	}, "Should panic because of non-existing base dir file")
}

func TestCliListDiskMigrations(t *testing.T) {
	action := listDiskMigrationsAction
	executeMigrator(&configFile, &action, &notVerbose, CreateConnector, CreateLoader)
}

func TestCliListDBTenants(t *testing.T) {
	action := listDBTenantsAction
	executeMigrator(&configFile, &action, &notVerbose, CreateConnector, CreateLoader)
}

func TestCliListDBMigrations(t *testing.T) {
	action := listDBMigrationsAction
	executeMigrator(&configFile, &action, &notVerbose, CreateConnector, CreateLoader)
}

func TestCliApply(t *testing.T) {
	action := applyAction
	executeMigrator(&configFile, &action, &verbose, CreateConnector, CreateLoader)
}
