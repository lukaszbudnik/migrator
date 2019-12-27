package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/types"
)

// will start returning errors when errorThreshold reached
type mockedErrorDiskLoader struct {
	errorThreshold int
	counter        int
}

func (m *mockedErrorDiskLoader) GetDiskMigrations() ([]types.Migration, error) {
	if m.errorThreshold == m.counter {
		return nil, fmt.Errorf("Mocked Error Disk Loader: threshold %v reached", m.errorThreshold)
	}
	m.counter++
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	m2 := types.Migration{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select def"}
	return []types.Migration{m1, m2}, nil
}

func newMockedErrorDiskLoader(errorThreshold int) func(config *config.Config) loader.Loader {
	return func(config *config.Config) loader.Loader {
		return &mockedErrorDiskLoader{errorThreshold: errorThreshold}
	}
}

func newMockedDiskLoader(config *config.Config) loader.Loader {
	return newMockedErrorDiskLoader(math.MaxInt64)(config)
}

type mockedBrokenCheckSumDiskLoader struct {
}

func (m *mockedBrokenCheckSumDiskLoader) GetDiskMigrations() ([]types.Migration, error) {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc", CheckSum: "xxx"}
	return []types.Migration{m1}, nil
}

func newBrokenCheckSumMockedDiskLoader(config *config.Config) loader.Loader {
	return new(mockedBrokenCheckSumDiskLoader)
}

// will start returning errors when errorThreshold reached
type mockedErrorConnector struct {
	errorThreshold int
	counter        int
}

func (m *mockedErrorConnector) Init() error {
	if m.errorThreshold == m.counter {
		return fmt.Errorf("Mocked Error Connector: threshold %v reached", m.errorThreshold)
	}
	m.counter++
	return nil
}

func (m *mockedErrorConnector) Dispose() {
}

func (m *mockedErrorConnector) AddTenantAndApplyMigrations(context.Context, string, []types.Migration) (*types.MigrationResults, error) {
	if m.errorThreshold == m.counter {
		return nil, fmt.Errorf("Mocked Error Connector: threshold %v reached", m.errorThreshold)
	}
	m.counter++
	return &types.MigrationResults{}, nil
}

func (m *mockedErrorConnector) GetTenants() ([]string, error) {
	if m.errorThreshold == m.counter {
		return nil, fmt.Errorf("Mocked Error Connector: threshold %v reached", m.errorThreshold)
	}
	m.counter++
	return []string{"a", "b", "c"}, nil
}

func (m *mockedErrorConnector) GetDBMigrations() ([]types.MigrationDB, error) {
	if m.errorThreshold == m.counter {
		return nil, fmt.Errorf("Mocked Error Connector: threshold %v reached", m.errorThreshold)
	}
	m.counter++
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.MigrationDB{{Migration: m1, Schema: "source", Created: d1}}
	return ms, nil
}

func (m *mockedErrorConnector) ApplyMigrations(ctx context.Context, migrations []types.Migration) (*types.MigrationResults, error) {
	if m.errorThreshold == m.counter {
		return nil, fmt.Errorf("Mocked Error Connector: threshold %v reached", m.errorThreshold)
	}
	m.counter++
	results := &types.MigrationResults{}
	return results, nil
}

func newMockedConnector(config *config.Config) (db.Connector, error) {
	return newMockedErrorConnector(math.MaxInt64)(config)
}

func newMockedErrorConnector(errorThreshold int) func(*config.Config) (db.Connector, error) {
	return func(config *config.Config) (db.Connector, error) {
		return &mockedErrorConnector{errorThreshold: errorThreshold}, nil
	}
}

func newConnectorReturnError(config *config.Config) (db.Connector, error) {
	return nil, errors.New("trouble maker")
}
