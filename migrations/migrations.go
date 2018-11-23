package migrations

import (
	"github.com/lukaszbudnik/migrator/types"
	"log"
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

// ComputeMigrationsToApply computes which disk migrations should be applied to DB based on migrations already present in DB
func ComputeMigrationsToApply(diskMigrations []types.Migration, dbMigrations []types.MigrationDB) []types.Migration {
	flattenedMigrationDBs := flattenMigrationDBs(dbMigrations)

	len := len(flattenedMigrationDBs)
	log.Printf("Number of flattened DB migrations: %d", len)

	var out []types.Migration

	for _, m := range diskMigrations[len:] {
		out = append(out, m)
	}

	return out
}
