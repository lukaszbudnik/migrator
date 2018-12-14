package core

import (
	"errors"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/types"
)

type mockedDiskLoader struct {
}

func (m *mockedDiskLoader) GetDiskMigrations() ([]types.Migration, error) {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema, Contents: "select abc", CheckSum: "abc"}
	m2 := types.Migration{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeTenantSchema, Contents: "select def", CheckSum: "def"}
	return []types.Migration{m1, m2}, nil
}

func createMockedDiskLoader(config *config.Config) loader.Loader {
	return new(mockedDiskLoader)
}

type mockedErrorDiskLoader struct {
}

func (m *mockedErrorDiskLoader) GetDiskMigrations() ([]types.Migration, error) {
	return []types.Migration{}, errors.New("disk trouble maker")
}

func createMockedErrorDiskLoader(config *config.Config) loader.Loader {
	return new(mockedErrorDiskLoader)
}

type mockedBrokenCheckSumDiskLoader struct {
}

func (m *mockedBrokenCheckSumDiskLoader) GetDiskMigrations() ([]types.Migration, error) {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema, Contents: "select abc", CheckSum: "xxx"}
	return []types.Migration{m1}, nil
}

func createBrokenCheckSumMockedDiskLoader(config *config.Config) loader.Loader {
	return new(mockedBrokenCheckSumDiskLoader)
}

type mockedConnector struct {
}

func (m *mockedConnector) Init() error {
	return nil
}

func (m *mockedConnector) Dispose() {
}

func (m *mockedConnector) GetTenants() ([]string, error) {
	return []string{"a", "b", "c"}, nil
}

func (m *mockedConnector) AddTenantAndApplyMigrations(string, []types.Migration) error {
	return nil
}

func (m *mockedConnector) GetDBMigrations() ([]types.MigrationDB, error) {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema, CheckSum: "abc"}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.MigrationDB{{Migration: m1, Schema: "source", Created: d1}}

	return ms, nil
}

func (m *mockedConnector) ApplyMigrations(migrations []types.Migration) error {
	return nil
}

func createMockedConnector(config *config.Config) db.Connector {
	return new(mockedConnector)
}

type mockedErrorConnector struct {
}

func (m *mockedErrorConnector) Init() error {
	return errors.New("trouble maker")
}

func (m *mockedErrorConnector) Dispose() {
}

func (m *mockedErrorConnector) AddTenantAndApplyMigrations(string, []types.Migration) error {
	return errors.New("trouble maker")
}

func (m *mockedErrorConnector) GetTenants() ([]string, error) {
	return []string{}, errors.New("trouble maker")
}

func (m *mockedErrorConnector) GetDBMigrations() ([]types.MigrationDB, error) {
	return []types.MigrationDB{}, errors.New("trouble maker")
}

func (m *mockedErrorConnector) ApplyMigrations(migrations []types.Migration) error {
	return errors.New("trouble maker")
}

func createMockedErrorConnector(config *config.Config) db.Connector {
	return new(mockedErrorConnector)
}

type mockedPassingVerificationErrorConnector struct {
}

func (m *mockedPassingVerificationErrorConnector) Init() error {
	return nil
}

func (m *mockedPassingVerificationErrorConnector) Dispose() {
}

func (m *mockedPassingVerificationErrorConnector) AddTenantAndApplyMigrations(string, []types.Migration) error {
	return errors.New("trouble maker")
}

func (m *mockedPassingVerificationErrorConnector) GetTenants() ([]string, error) {
	return []string{}, errors.New("trouble maker")
}

func (m *mockedPassingVerificationErrorConnector) GetDBMigrations() ([]types.MigrationDB, error) {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema, CheckSum: "abc"}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.MigrationDB{{Migration: m1, Schema: "source", Created: d1}}

	return ms, nil
}

func (m *mockedPassingVerificationErrorConnector) ApplyMigrations(migrations []types.Migration) error {
	return errors.New("trouble maker")
}

func createMockedPassingVerificationErrorConnector(config *config.Config) db.Connector {
	return new(mockedPassingVerificationErrorConnector)
}
