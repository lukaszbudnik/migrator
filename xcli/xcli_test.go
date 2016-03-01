// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * disk_test.go
// * migrations_test.go

package xcli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	unknownAction = "unknown"
	configFile    = "../test/migrator.yaml"
	verbose       = true
	notVerbose    = false
)

func TestCliExitUnknownAction(t *testing.T) {
	ret := ExecuteMigrator(&configFile, &unknownAction, &notVerbose, nil, nil)
	assert.Equal(t, 1, ret)
}

func TestCliListDiskMigrations(t *testing.T) {
	action := ListDiskMigrationsAction
	ExecuteMigrator(&configFile, &action, &notVerbose, nil, createMockedDiskLoader)
}

func TestCliListDBTenants(t *testing.T) {
	action := ListDBTenantsAction
	ExecuteMigrator(&configFile, &action, &notVerbose, createMockedConnector, nil)
}

func TestCliListMigrationDBs(t *testing.T) {
	action := ListDBMigrationsAction
	ExecuteMigrator(&configFile, &action, &notVerbose, createMockedConnector, nil)
}

func TestCliApply(t *testing.T) {
	action := ApplyAction
	ExecuteMigrator(&configFile, &action, &verbose, createMockedConnector, createMockedDiskLoader)
}
