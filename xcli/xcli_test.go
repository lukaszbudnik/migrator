// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * disk_test.go
// * migrations_test.go

package xcli

import (
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/disk"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	unknownAction                    = "unknown"
	nonExistingConfig                = "idontexist"
	configFileWithNonExistingBaseDir = "../test/migrator-test-non-existing-base-dir.yaml"
	configFile                       = "../test/migrator.yaml"
	verbose                          = true
	notVerbose                       = false
)

func TestCliExitUnknownAction(t *testing.T) {
	ret := ExecuteMigrator(&nonExistingConfig, &unknownAction, &notVerbose, db.CreateConnector, disk.CreateLoader)
	assert.Equal(t, 1, ret)
}

func TestCliPanicReadFromNonExistingConfigFile(t *testing.T) {
	action := ApplyAction
	assert.Panics(t, func() {
		ExecuteMigrator(&nonExistingConfig, &action, &notVerbose, db.CreateConnector, disk.CreateLoader)
	}, "Should panic because of non-existing config file")
}

func TestCliPanicReadDiskMigrationsFromNonExistingBaseDir(t *testing.T) {
	action := ApplyAction
	assert.Panics(t, func() {
		ExecuteMigrator(&configFileWithNonExistingBaseDir, &action, &notVerbose, db.CreateConnector, disk.CreateLoader)
	}, "Should panic because of non-existing base dir file")
}

func TestCliListDiskMigrations(t *testing.T) {
	action := ListDiskMigrationsAction
	ExecuteMigrator(&configFile, &action, &notVerbose, db.CreateConnector, disk.CreateLoader)
}

func TestCliListDBTenants(t *testing.T) {
	action := ListDBTenantsAction
	ExecuteMigrator(&configFile, &action, &notVerbose, db.CreateConnector, disk.CreateLoader)
}

func TestCliListMigrationDBs(t *testing.T) {
	action := ListMigrationDBsAction
	ExecuteMigrator(&configFile, &action, &notVerbose, db.CreateConnector, disk.CreateLoader)
}

func TestCliApply(t *testing.T) {
	action := ApplyAction
	ExecuteMigrator(&configFile, &action, &verbose, db.CreateConnector, disk.CreateLoader)
}
