package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	unknownAction                    = "unknown"
	defaultAction                    = "apply"
	nonExistingConfig                = "idontexist"
	configFileWithNonExistingBaseDir = "test/migrator-test-non-existing-base-dir.yaml"
	verbose                          = false
)

func TestCliExitUnknownAction(t *testing.T) {
	ret := executeMigrator(&nonExistingConfig, &unknownAction, &verbose, CreateConnector)
	assert.Equal(t, 1, ret)
}

func TestCliPanicReadFromNonExistingConfigFile(t *testing.T) {
	assert.Panics(t, func() {
		executeMigrator(&nonExistingConfig, &defaultAction, &verbose, CreateConnector)
	}, "Should panic because of non-existing config file")
}

func TestCliPanicReadDiskMigrationsFromNonExistingBaseDir(t *testing.T) {
	assert.Panics(t, func() {
		executeMigrator(&configFileWithNonExistingBaseDir, &defaultAction, &verbose, CreateConnector)
	}, "Should panic because of non-existing base dir file")
}
