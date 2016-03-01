package migrations

import (
	"github.com/lukaszbudnik/migrator/types"
)

func flattenDBMigrations(dbMigrations []types.DBMigration) []types.MigrationDefinition {
	var flattened []types.MigrationDefinition
	var previousMigrationDefinition types.MigrationDefinition
	for i, m := range dbMigrations {
		if i == 0 || m.MigrationType == types.ModeSingleSchema || m.MigrationDefinition != previousMigrationDefinition {
			flattened = append(flattened, m.MigrationDefinition)
			previousMigrationDefinition = m.MigrationDefinition
		}
	}
	return flattened
}

func ComputeMigrationsToApply(diskMigrations []types.Migration, dbMigrations []types.DBMigration) []types.Migration {
	flattenedDBMigrations := flattenDBMigrations(dbMigrations)

	len := len(flattenedDBMigrations)
	var out []types.Migration

	for _, m := range diskMigrations[len:] {
		out = append(out, m)
	}

	return out
}
