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

// difference returns the elements on disk which are not yet in DB
func difference(diskMigrations []types.Migration, flattenedMigrationDBs []types.MigrationDefinition) []types.Migration {
		// key is Migration.File
    existsInDB := map[string]bool{}
    for _, m := range flattenedMigrationDBs {
        existsInDB[m.File] = true
    }
    diff := []types.Migration{}
    for _, m := range diskMigrations {
        if _, ok := existsInDB[m.File]; !ok {
            diff = append(diff, m)
        }
    }
    return diff
}

// ComputeMigrationsToApply computes which disk migrations should be applied to DB based on migrations already present in DB
func ComputeMigrationsToApply(diskMigrations []types.Migration, dbMigrations []types.MigrationDB) []types.Migration {
	flattenedMigrationDBs := flattenMigrationDBs(dbMigrations)

	len := len(flattenedMigrationDBs)
	log.Printf("Number of flattened DB migrations: %d", len)

	out := difference(diskMigrations, flattenedMigrationDBs)

	return out
}
