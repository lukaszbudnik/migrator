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

func TestLoadDiskMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	LoadDiskMigrations(config, createMockedDiskLoader)
}

func TestLoadDBTenants(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	LoadDBTenants(config, createMockedConnector)
}

func TestLoadDBMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	LoadDBMigrations(config, createMockedConnector)
}

func TestApplyMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	ApplyMigrations(config, createMockedConnector, createMockedDiskLoader)
}
