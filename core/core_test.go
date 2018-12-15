package core

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
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
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
	assert.Nil(t, err)
}

func TestGetDiskMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDiskMigrationsAction
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
	assert.Nil(t, err)
}

func TestGetDiskMigrationsError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDiskMigrationsAction
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedErrorDiskLoader)
	assert.Equal(t, "disk trouble maker", err.Error())
}

func TestGetDBTenants(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDBTenantsAction
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
	assert.Nil(t, err)
}

func TestGetDBMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDBMigrationsAction
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
	assert.Nil(t, err)
}

func TestApplyMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = ApplyAction
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
	assert.Nil(t, err)
}

func TestApplyMigrationsVerificationFailed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = ApplyAction
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createBrokenCheckSumMockedDiskLoader)
	assert.Equal(t, "Checksum verification failed", err.Error())
}

func TestAddTenant(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = AddTenantAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedDiskLoader)
}

func TestAddTenantVerificationFailed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = AddTenantAction
	doExecuteMigrator(config, executeFlags, createMockedConnector, createBrokenCheckSumMockedDiskLoader)
}

func TestGetDBTenantsError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDBTenantsAction
	err = doExecuteMigrator(config, executeFlags, createMockedErrorConnector, createMockedDiskLoader)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestGetDBMigrationsError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = GetDBMigrationsAction
	err = doExecuteMigrator(config, executeFlags, createMockedErrorConnector, createMockedDiskLoader)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestApplyMigrationsDBError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = ApplyAction
	err = doExecuteMigrator(config, executeFlags, createMockedErrorConnector, createMockedDiskLoader)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestApplyMigrationsDirectDBError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	_, err = ApplyMigrations(config, createMockedErrorConnector, createMockedDiskLoader)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestApplyMigrationsDiskError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = ApplyAction
	err = doExecuteMigrator(config, executeFlags, createMockedConnector, createMockedErrorDiskLoader)
	assert.Equal(t, "disk trouble maker", err.Error())
}

func TestApplyMigrationsDirectDiskError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	_, err = ApplyMigrations(config, createMockedConnector, createMockedErrorDiskLoader)
	assert.Equal(t, "disk trouble maker", err.Error())
}

func TestApplyMigrationsPassingVerificationError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = ApplyAction
	err = doExecuteMigrator(config, executeFlags, createMockedPassingVerificationErrorConnector, createMockedDiskLoader)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestDoApplyMigrationsDBError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = ApplyAction
	err = doApplyMigrations([]types.Migration{}, config, createMockedErrorConnector)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestAddTenantDBError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = AddTenantAction
	err = doExecuteMigrator(config, executeFlags, createMockedErrorConnector, createMockedDiskLoader)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestAddTenantDirectDiskError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	_, err = AddTenant("tenant", config, createMockedConnector, createMockedErrorDiskLoader)
	assert.Equal(t, "disk trouble maker", err.Error())
}

func TestAddTenantPassingVerificationError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = AddTenantAction
	err = doExecuteMigrator(config, executeFlags, createMockedPassingVerificationErrorConnector, createMockedDiskLoader)
	assert.Equal(t, "trouble maker", err.Error())
}

func TestDoAddTenantDBError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	executeFlags := ExecuteFlags{}
	executeFlags.Action = AddTenantAction
	err = doAddTenantAndApplyMigrations("tencio", []types.Migration{}, config, createMockedErrorConnector)
	assert.Equal(t, "trouble maker", err.Error())
}
