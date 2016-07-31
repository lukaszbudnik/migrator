// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * migrations_test.go
// DB & Disk operations are mocked using xcli_mocks.go

package xcli

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	unknownAction = "unknown"
	configFile    = "../test/migrator.yaml"
)

func TestCliExitUnknownAction(t *testing.T) {
	config := config.FromFile(configFile)
	ret := ExecuteMigrator(config, &unknownAction, nil, nil)
	assert.Equal(t, 1, ret)
}

func TestCliListDiskMigrations(t *testing.T) {
	config := config.FromFile(configFile)
	action := ListDiskMigrationsAction
	ExecuteMigrator(config, &action, nil, createMockedDiskLoader)
}

func TestCliListDBTenants(t *testing.T) {
	config := config.FromFile(configFile)
	action := ListDBTenantsAction
	ExecuteMigrator(config, &action, createMockedConnector, nil)
}

func TestCliListMigrationDBs(t *testing.T) {
	config := config.FromFile(configFile)
	action := ListDBMigrationsAction
	ExecuteMigrator(config, &action, createMockedConnector, nil)
}

func TestCliApply(t *testing.T) {
	config := config.FromFile(configFile)
	action := ApplyAction
	ExecuteMigrator(config, &action, createMockedConnector, createMockedDiskLoader)
}

func TestCliReadConfig(t *testing.T) {
	config := ReadConfig(&configFile)
	assert.NotNil(t, config)
}

func TestCliPrintConfig(t *testing.T) {
	config := config.FromFile(configFile)
	action := PrintConfigAction
	ExecuteMigrator(config, &action, nil, nil)
}
