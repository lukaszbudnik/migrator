package migrations

import (
	"context"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/types"
)

func flattenMigrationDBs(dbMigrations []types.MigrationDB) []types.Migration {
	var flattened []types.Migration
	var previousMigration types.Migration
	for i, m := range dbMigrations {
		if i == 0 || m.Migration != previousMigration {
			flattened = append(flattened, m.Migration)
			previousMigration = m.Migration
		}
	}
	return flattened
}

// difference returns the elements on disk which are not yet in DB
// the exceptions are MigrationTypeSingleScript and MigrationTypeTenantScript which are always run
func difference(diskMigrations []types.Migration, flattenedMigrationDBs []types.Migration) []types.Migration {
	// key is Migration.File
	existsInDB := map[string]bool{}
	for _, m := range flattenedMigrationDBs {
		if m.MigrationType != types.MigrationTypeSingleScript && m.MigrationType != types.MigrationTypeTenantScript {
			existsInDB[m.File] = true
		}
	}
	diff := []types.Migration{}
	for _, m := range diskMigrations {
		if _, ok := existsInDB[m.File]; !ok {
			diff = append(diff, m)
		}
	}
	return diff
}

// intersect returns the elements on disk and in DB
func intersect(diskMigrations []types.Migration, flattenedMigrationDBs []types.Migration) []struct {
	disk types.Migration
	db   types.Migration
} {
	// key is Migration.File
	existsInDB := map[string]types.Migration{}
	for _, m := range flattenedMigrationDBs {
		existsInDB[m.File] = m
	}
	intersect := []struct {
		disk types.Migration
		db   types.Migration
	}{}
	for _, m := range diskMigrations {
		if db, ok := existsInDB[m.File]; ok {
			intersect = append(intersect, struct {
				disk types.Migration
				db   types.Migration
			}{m, db})
		}
	}
	return intersect
}

// ComputeMigrationsToApply computes which disk migrations should be applied to DB based on migrations already present in DB
func ComputeMigrationsToApply(ctx context.Context, diskMigrations []types.Migration, dbMigrations []types.MigrationDB) []types.Migration {
	flattenedMigrationDBs := flattenMigrationDBs(dbMigrations)

	len := len(flattenedMigrationDBs)
	common.LogInfo(ctx, "Number of flattened DB migrations: %d", len)

	out := difference(diskMigrations, flattenedMigrationDBs)

	return out
}

// FilterTenantMigrations returns only migrations which are of type MigrationTypeTenantSchema
func FilterTenantMigrations(ctx context.Context, diskMigrations []types.Migration) []types.Migration {
	filteredTenantMigrations := []types.Migration{}
	for _, m := range diskMigrations {
		if m.MigrationType == types.MigrationTypeTenantMigration {
			filteredTenantMigrations = append(filteredTenantMigrations, m)
		}
	}

	len := len(filteredTenantMigrations)
	common.LogInfo(ctx, "Number of filtered tenant DB migrations: %d", len)

	return filteredTenantMigrations
}

// VerifyCheckSums verifies if CheckSum of disk and flattened DB migrations match
// returns bool indicating if offending (i.e., modified) disk migrations were found
// if bool is false the function returns a slice of offending migrations
// if bool is true the slice of effending migrations is empty
func VerifyCheckSums(diskMigrations []types.Migration, dbMigrations []types.MigrationDB) (bool, []types.Migration) {

	flattenedMigrationDBs := flattenMigrationDBs(dbMigrations)

	intersect := intersect(diskMigrations, flattenedMigrationDBs)
	var offendingMigrations []types.Migration
	var result = true
	for _, t := range intersect {
		if t.disk.CheckSum != t.db.CheckSum {
			offendingMigrations = append(offendingMigrations, t.disk)
			result = false
		}
	}
	return result, offendingMigrations
}
