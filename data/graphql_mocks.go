package data

import (
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/types"
)

type mockedCoordinator struct {
}

func (m *mockedCoordinator) safeString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (m *mockedCoordinator) CreateTenant(string, types.Action, bool, string) *types.CreateResults {
	version, _ := m.GetVersionByID(0)
	return &types.CreateResults{Summary: &types.Summary{}, Version: version}
}

func (m *mockedCoordinator) CreateVersion(string, types.Action, bool) *types.CreateResults {
	// re-use mocked version from GetVersionByID...
	version, _ := m.GetVersionByID(0)
	return &types.CreateResults{Summary: &types.Summary{}, Version: version}
}

func (m *mockedCoordinator) GetSourceMigrations(filters *coordinator.SourceMigrationFilters) []types.Migration {

	if filters == nil {
		m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
		m2 := types.Migration{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select def"}
		m3 := types.Migration{Name: "201602220001.sql", SourceDir: "config", File: "config/201602220001.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select def"}
		m4 := types.Migration{Name: "201602220002.sql", SourceDir: "source", File: "source/201602220002.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select def"}
		m5 := types.Migration{Name: "201602220003.sql", SourceDir: "tenant", File: "tenant/201602220003.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select def"}
		return []types.Migration{m1, m2, m3, m4, m5}
	}

	m1 := types.Migration{Name: m.safeString(filters.Name), SourceDir: m.safeString(filters.SourceDir), File: m.safeString(filters.File), MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	return []types.Migration{m1}
}

func (m *mockedCoordinator) GetSourceMigrationByFile(file string) (*types.Migration, error) {
	i := strings.Index(file, "/")
	sourceDir := file[:i]
	name := file[i+1:]
	m1 := types.Migration{Name: name, SourceDir: sourceDir, File: file, MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	return &m1, nil
}

func (m *mockedCoordinator) Dispose() {
}

func (m *mockedCoordinator) GetTenants() []types.Tenant {
	a := types.Tenant{Name: "a"}
	b := types.Tenant{Name: "b"}
	c := types.Tenant{Name: "c"}
	return []types.Tenant{a, b, c}
}

func (m *mockedCoordinator) GetVersions() []types.Version {
	a := types.Version{ID: 12, Name: "a", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -2)}}
	b := types.Version{ID: 121, Name: "bb", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -1)}}
	c := types.Version{ID: 122, Name: "ccc", Created: graphql.Time{Time: time.Now()}}
	return []types.Version{a, b, c}
}

func (m *mockedCoordinator) GetVersionsByFile(file string) []types.Version {
	a := types.Version{ID: 12, Name: "a", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -2)}}
	return []types.Version{a}
}

func (m *mockedCoordinator) GetVersionByID(ID int32) (*types.Version, error) {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	db1 := types.DBMigration{Migration: m1, Schema: "source", Created: graphql.Time{Time: d1}}

	m2 := types.Migration{Name: "202002180000.sql", SourceDir: "config", File: "config/202002180000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	d2 := time.Date(2020, 02, 18, 16, 41, 1, 123, time.UTC)
	db2 := types.DBMigration{Migration: m2, Schema: "source", Created: graphql.Time{Time: d2}}

	m3 := types.Migration{Name: "202002180000.sql", SourceDir: "tenants", File: "tenants/202002180000.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select abc"}
	d3 := time.Date(2020, 02, 18, 16, 41, 1, 123, time.UTC)
	db3 := types.DBMigration{Migration: m3, Schema: "abc", Created: graphql.Time{Time: d3}}
	db4 := types.DBMigration{Migration: m3, Schema: "def", Created: graphql.Time{Time: d3}}
	db5 := types.DBMigration{Migration: m3, Schema: "xyz", Created: graphql.Time{Time: d3}}

	a := types.Version{ID: ID, Name: "a", Created: graphql.Time{Time: time.Now().AddDate(0, 0, -2)}, DBMigrations: []types.DBMigration{db1, db2, db3, db4, db5}}

	return &a, nil
}

// not used in GraphQL
func (m *mockedCoordinator) GetAppliedMigrations() []types.DBMigration {
	return []types.DBMigration{}
}

func (m *mockedCoordinator) GetDBMigrationByID(ID int32) (*types.DBMigration, error) {
	migration := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	d := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	db := types.DBMigration{Migration: migration, ID: ID, Schema: "source", Created: graphql.Time{Time: d}}
	return &db, nil
}

func (m *mockedCoordinator) VerifySourceMigrationsCheckSums() (bool, []types.Migration) {
	return true, nil
}
