package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/metrics"
	"github.com/lukaszbudnik/migrator/types"
)

type mockedCoordinator struct {
	errorThreshold int
	counter        int
}

func newMockedCoordinator(ctx context.Context, config *config.Config, metrics metrics.Metrics) coordinator.Coordinator {
	return newMockedErrorCoordinator(-1)(ctx, config, metrics)
}

func newMockedErrorCoordinator(errorThreshold int) func(context.Context, *config.Config, metrics.Metrics) coordinator.Coordinator {
	return func(ctx context.Context, config *config.Config, metrics metrics.Metrics) coordinator.Coordinator {
		return &mockedCoordinator{errorThreshold: errorThreshold}
	}
}

func (m *mockedCoordinator) Dispose() {
}

func (m *mockedCoordinator) CreateTenant(string, types.Action, bool, string) *types.CreateResults {
	return &types.CreateResults{Summary: &types.Summary{}, Version: &types.Version{}}
}

func (m *mockedCoordinator) CreateVersion(string, types.Action, bool) *types.CreateResults {
	return &types.CreateResults{Summary: &types.Summary{}, Version: &types.Version{}}
}

func (m *mockedCoordinator) GetSourceMigrations(_ *coordinator.SourceMigrationFilters) []types.Migration {
	if m.errorThreshold == m.counter {
		panic(fmt.Sprintf("Mocked Coordinator: threshold %v reached", m.errorThreshold))
	}
	if m.errorThreshold != -1 {
		m.counter++
	}
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	m2 := types.Migration{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select def"}
	return []types.Migration{m1, m2}
}

func (m *mockedCoordinator) GetSourceMigrationByFile(file string) (*types.Migration, error) {
	if m.errorThreshold == m.counter {
		panic(fmt.Sprintf("Mocked Coordinator: threshold %v reached", m.errorThreshold))
	}
	if m.errorThreshold != -1 {
		m.counter++
	}
	i := strings.Index(file, "/")
	sourceDir := file[:i]
	name := file[i+1:]
	m1 := types.Migration{Name: name, SourceDir: sourceDir, File: file, MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	return &m1, nil
}

func (m *mockedCoordinator) GetAppliedMigrations() []types.DBMigration {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc", CheckSum: "sha256"}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.DBMigration{{Migration: m1, Schema: "source", Created: graphql.Time{Time: d1}}}
	return ms
}

// part of interface but not used in server tests - tested in data package
func (m *mockedCoordinator) GetDBMigrationByID(ID int32) (*types.DBMigration, error) {
	return nil, nil
}

func (m *mockedCoordinator) GetTenants() []types.Tenant {
	a := types.Tenant{Name: "a"}
	b := types.Tenant{Name: "b"}
	c := types.Tenant{Name: "c"}
	return []types.Tenant{a, b, c}
}

// part of interface but not used in server tests - tested in data package
func (m *mockedCoordinator) GetVersions() []types.Version {
	return []types.Version{}
}

// part of interface but not used in server tests - tested in data package
func (m *mockedCoordinator) GetVersionsByFile(file string) []types.Version {
	return []types.Version{}
}

// part of interface but not used in server tests - tested in data package
func (m *mockedCoordinator) GetVersionByID(ID int32) (*types.Version, error) {
	return nil, nil
}

func (m *mockedCoordinator) VerifySourceMigrationsCheckSums() (bool, []types.Migration) {
	if m.errorThreshold == m.counter {
		m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc", CheckSum: "123"}
		return false, []types.Migration{m1}
	}
	m.counter++
	return true, nil
}

func (m *mockedCoordinator) HealthCheck() types.HealthResponse {
	if m.errorThreshold == m.counter {
		panic(fmt.Sprintf("Mocked Coordinator: threshold %v reached", m.errorThreshold))
	}
	if m.errorThreshold != -1 {
		m.counter++
	}
	return types.HealthResponse{Status: types.HealthStatusUp, Checks: []types.HealthChecks{}}
}

type mockedCoordinatorHealthCheckError struct {
	mockedCoordinator
}

func (m *mockedCoordinatorHealthCheckError) HealthCheck() types.HealthResponse {
	return types.HealthResponse{Status: types.HealthStatusDown, Checks: []types.HealthChecks{}}
}

func newMockedCoordinatorHealthCheckError(ctx context.Context, config *config.Config, metrics metrics.Metrics) coordinator.Coordinator {
	return &mockedCoordinatorHealthCheckError{}
}

func newNoopMetrics() metrics.Metrics {
	return &noopMetrics{}
}

type noopMetrics struct {
}

func (m *noopMetrics) SetGaugeValue(name string, labelValues []string, value float64) error {
	return nil
}

func (m *noopMetrics) AddGaugeValue(name string, labelValues []string, value float64) error {
	return nil
}

func (m *noopMetrics) IncrementGaugeValue(name string, labelValues []string) error {
	return nil
}
