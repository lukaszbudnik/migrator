package migrations

import (
	"github.com/lukaszbudnik/migrator/types"
)

func flattenMigrationDBs(dbMigrations []types.MigrationDB) []types.MigrationDefinition {
	var flattened []types.MigrationDefinition
	var previousMigrationDefinition types.MigrationDefinition
	for i, m := range dbMigrations {
		if i == 0 || m.MigrationType == types.MigrationTypeSingleSchema || m.MigrationDefinition != previousMigrationDefinition {
			flattened = append(flattened, m.MigrationDefinition)
			previousMigrationDefinition = m.MigrationDefinition
		}
	}
	return flattened
}

func ComputeMigrationsToApply(diskMigrations []types.Migration, dbMigrations []types.MigrationDB) []types.Migration {
	flattenedMigrationDBs := flattenMigrationDBs(dbMigrations)

	len := len(flattenedMigrationDBs)
	var out []types.Migration

	for _, m := range diskMigrations[len:] {
		out = append(out, m)
	}

	return out
}
