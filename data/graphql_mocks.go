package data

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/lukaszbudnik/migrator/types"
)

type mockedCoordinator struct {
}

func (m *mockedCoordinator) GetSourceMigrations() []types.Migration {

	// 5 migrations in total
	// 4 migrations with type MigrationTypeSingleMigration
	// 3 migrations with sourceDir source and type MigrationTypeSingleMigration
	// 2 migrations with name 201602220001.sql and type MigrationTypeSingleMigration
	// 1 migration with file config/201602220001.sql

	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	m2 := types.Migration{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select def"}
	m3 := types.Migration{Name: "201602220001.sql", SourceDir: "config", File: "config/201602220001.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select def"}
	m4 := types.Migration{Name: "201602220002.sql", SourceDir: "source", File: "source/201602220002.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select def"}
	m5 := types.Migration{Name: "201602220003.sql", SourceDir: "tenant", File: "tenant/201602220003.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select def"}
	return []types.Migration{m1, m2, m3, m4, m5}
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

func (m *mockedCoordinator) GetAppliedMigrations() []types.MigrationDB {
	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	db1 := types.MigrationDB{Migration: m1, Schema: "source", Created: graphql.Time{Time: d1}}

	m2 := types.Migration{Name: "202002180000.sql", SourceDir: "config", File: "config/202002180000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc"}
	d2 := time.Date(2020, 02, 18, 16, 41, 1, 123, time.UTC)
	db2 := types.MigrationDB{Migration: m2, Schema: "source", Created: graphql.Time{Time: d2}}

	m3 := types.Migration{Name: "202002180000.sql", SourceDir: "tenants", File: "tenants/202002180000.sql", MigrationType: types.MigrationTypeTenantMigration, Contents: "select abc"}
	d3 := time.Date(2020, 02, 18, 16, 41, 1, 123, time.UTC)
	db3 := types.MigrationDB{Migration: m3, Schema: "abc", Created: graphql.Time{Time: d3}}
	db4 := types.MigrationDB{Migration: m3, Schema: "def", Created: graphql.Time{Time: d3}}
	db5 := types.MigrationDB{Migration: m3, Schema: "xyz", Created: graphql.Time{Time: d3}}

	return []types.MigrationDB{db1, db2, db3, db4, db5}
}

func (m *mockedCoordinator) ApplyMigrations(types.MigrationsModeType) (*types.MigrationResults, []types.Migration) {
	return &types.MigrationResults{}, []types.Migration{}
}

func (m *mockedCoordinator) AddTenantAndApplyMigrations(types.MigrationsModeType, string) (*types.MigrationResults, []types.Migration) {
	return &types.MigrationResults{}, []types.Migration{}
}

func (m *mockedCoordinator) VerifySourceMigrationsCheckSums() (bool, []types.Migration) {
	return true, nil
}
