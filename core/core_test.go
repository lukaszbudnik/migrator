// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * migrations_test.go
// DB & Disk operations are mocked using xcli_mocks.go

package core

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	unknownAction = "unknown"
	configFile    = "../test/migrator.yaml"
)

func TestPrintConfig(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = PrintConfigAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
}

func TestGetDiskMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDiskMigrationsAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
}

func TestGetDBTenants(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDBTenantsAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
}

func TestGetDBMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDBMigrationsAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
}

func TestApplyMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = ApplyAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
}

func TestAddTenant(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = AddTenantAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
}
