package coordinator

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
)

type mockedDiskLoader struct {
}

func (m *mockedDiskLoader) GetSourceMigrations() []types.Migration {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	m2 := types.Migration{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select def"}
	return []types.Migration{m1, m2}
}

func newMockedDiskLoader(_ context.Context, _ *config.Config) loader.Loader {
	return &mockedDiskLoader{}
}

type mockedNotifier struct{}

func (m *mockedNotifier) Notify(message string) (string, error) {
	return "mock", nil
}

func newMockedNotifier(_ context.Context, _ *config.Config) notifications.Notifier {
	return &mockedNotifier{}
}

type mockedBrokenCheckSumDiskLoader struct {
}

func (m *mockedBrokenCheckSumDiskLoader) GetSourceMigrations() []types.Migration {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc", CheckSum: "xxx"}
	return []types.Migration{m1}
}

func newBrokenCheckSumMockedDiskLoader(_ context.Context, _ *config.Config) loader.Loader {
	return new(mockedBrokenCheckSumDiskLoader)
}

type mockedDifferentScriptCheckSumMockedDiskLoader struct {
}

func (m *mockedDifferentScriptCheckSumMockedDiskLoader) GetSourceMigrations() []types.Migration {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	m2 := types.Migration{Name: "recreate-indexes.sql", SourceDir: "tenants-scripts", File: "tenants-scripts/recreate-indexes.sql", MigrationType: types.MigrationTypeTenantScript, Contents: "select abc", CheckSum: "sha256-1"}
	return []types.Migration{m1, m2}
}

func newDifferentScriptCheckSumMockedDiskLoader(_ context.Context, _ *config.Config) loader.Loader {
	return new(mockedDifferentScriptCheckSumMockedDiskLoader)
}

type mockedConnector struct {
}

func (m *mockedConnector) Dispose() {
}

func (m *mockedConnector) AddTenantAndApplyMigrations(types.MigrationsModeType, string, []types.Migration) *types.MigrationResults {
	return &types.MigrationResults{}
}

func (m *mockedConnector) GetTenants() []types.Tenant {
	a := types.Tenant{Name: "a"}
	b := types.Tenant{Name: "b"}
	c := types.Tenant{Name: "c"}
	return []types.Tenant{a, b, c}
}

func (m *mockedConnector) GetVersions() []types.Version {
	a := types.Version{ID: 12, Name: "a", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -2)}}
	b := types.Version{ID: 121, Name: "bb", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -1)}}
	c := types.Version{ID: 122, Name: "ccc", Created: graphql.Time{Time: time.Now()}}
	return []types.Version{a, b, c}
}

func (m *mockedConnector) GetVersionsByFile(file string) []types.Version {
	a := types.Version{ID: 12, Name: "a", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -2)}}
	return []types.Version{a}
}

func (m *mockedConnector) GetVersionByID(ID int32) types.Version {
	a := types.Version{ID: ID, Name: "a", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -2)}}
	return a
}

func (m *mockedConnector) GetAppliedMigrations() []types.MigrationDB {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.MigrationDB{{Migration: m1, Schema: "source", AppliedAt: graphql.Time{Time: d1}}}
	return ms
}

func (m *mockedConnector) ApplyMigrations(types.MigrationsModeType, []types.Migration) *types.MigrationResults {
	return &types.MigrationResults{}
}

func newMockedConnector(context.Context, *config.Config) db.Connector {
	return &mockedConnector{}
}

type mockedDifferentScriptCheckSumMockedConnector struct {
	mockedConnector
}

func (m *mockedDifferentScriptCheckSumMockedConnector) GetAppliedMigrations() []types.MigrationDB {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	m2 := types.Migration{Name: "recreate-indexes.sql", SourceDir: "tenants-scripts", File: "tenants-scripts/recreate-indexes.sql", MigrationType: types.MigrationTypeTenantScript, Contents: "select abc", CheckSum: "sha256-2"}
	d2 := time.Date(2016, 02, 22, 16, 41, 1, 456, time.UTC)
	ms := []types.MigrationDB{{Migration: m1, Schema: "source", AppliedAt: graphql.Time{Time: d1}}, {Migration: m2, Schema: "customer1", AppliedAt: graphql.Time{Time: d2}}}
	return ms
}

func newDifferentScriptCheckSumMockedConnector(context.Context, *config.Config) db.Connector {
	return &mockedDifferentScriptCheckSumMockedConnector{mockedConnector{}}
}
