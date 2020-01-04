package server

import (
	"context"
	"fmt"
	"time"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/types"
)

type mockedCoordinator struct {
	errorThreshold int
	counter        int
}

func newMockedCoordinator(ctx context.Context, config *config.Config) coordinator.Coordinator {
	return newMockedErrorCoordinator(-1)(ctx, config)
}

func newMockedErrorCoordinator(errorThreshold int) func(context.Context, *config.Config) coordinator.Coordinator {
	return func(ctx context.Context, config *config.Config) coordinator.Coordinator {
		return &mockedCoordinator{errorThreshold: errorThreshold}
	}
}

func (m *mockedCoordinator) Dispose() {
}

func (m *mockedCoordinator) GetSourceMigrations() []types.Migration {
	if m.errorThreshold == m.counter {
		panic(fmt.Sprintf("Mocked Error Disk Loader: threshold %v reached", m.errorThreshold))
	}
	m.counter++
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	m2 := types.Migration{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select def"}
	return []types.Migration{m1, m2}
}

func (m *mockedCoordinator) GetAppliedMigrations() []types.MigrationDB {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.MigrationDB{{Migration: m1, Schema: "source", AppliedAt: d1}}
	return ms
}

func (m *mockedCoordinator) GetTenants() []string {
	return []string{"a", "b", "c"}
}

func (m *mockedCoordinator) VerifySourceMigrationsCheckSums() (bool, []types.Migration) {
	if m.errorThreshold == m.counter {
		m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc", CheckSum: "123"}
		return false, []types.Migration{m1}
	}
	m.counter++
	return true, nil
}

func (m *mockedCoordinator) ApplyMigrations(types.MigrationsModeType) (*types.MigrationResults, []types.Migration) {
	return &types.MigrationResults{}, m.GetSourceMigrations()
}

func (m *mockedCoordinator) AddTenantAndApplyMigrations(types.MigrationsModeType, string) (*types.MigrationResults, []types.Migration) {
	return &types.MigrationResults{}, m.GetSourceMigrations()[1:]
}
