package data

import (
	"github.com/lukaszbudnik/migrator/types"
)

type mockedLoader struct {
}

func (m *mockedLoader) GetSourceMigrations() []types.Migration {

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

// func (m *mockedCoordinator) GetAppliedMigrations() []types.MigrationDB {
// 	m1 := types.Migration{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleMigration, Contents: "select abc", CheckSum: "sha256"}
// 	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
// 	ms := []types.MigrationDB{{Migration: m1, Schema: "source", AppliedAt: d1}}
// 	return ms
// }
// func (m *mockedCoordinator) GetTenants() []string {
// 	return []string{"a", "b", "c"}
// }
