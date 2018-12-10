package migrations

import (
	"log"

	"github.com/lukaszbudnik/migrator/types"
)

func flattenMigrationDBs(dbMigrations []types.MigrationDB) []types.Migration {
	var flattened []types.Migration
	var previousMigration types.Migration
	for i, m := range dbMigrations {
		if i == 0 || m.MigrationType == types.MigrationTypeSingleSchema || m.Migration != previousMigration {
			flattened = append(flattened, m.Migration)
			previousMigration = m.Migration
		}
	}
	return flattened
}

// difference returns the elements on disk which are not yet in DB
func difference(diskMigrations []types.Migration, flattenedMigrationDBs []types.Migration) []types.Migration {
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

// FilterTenantMigrations returns only migrations which are of type MigrationTypeTenantSchema
func FilterTenantMigrations(diskMigrations []types.Migration) []types.Migration {
	filteredTenantMigrations := []types.Migration{}
	for _, m := range diskMigrations {
		if m.MigrationType == types.MigrationTypeTenantSchema {
			filteredTenantMigrations = append(filteredTenantMigrations, m)
		}
	}

	len := len(filteredTenantMigrations)
	log.Printf("Number of flattened DB migrations: %d", len)

	return filteredTenantMigrations
}
