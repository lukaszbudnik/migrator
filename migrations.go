package main

import "log"

func flattenDBMigrations(dbMigrations []DBMigration) []MigrationDefinition {
	var flattened []MigrationDefinition
	var previousMigrationDefinition MigrationDefinition
	for i, m := range dbMigrations {
		if i == 0 || m.MigrationType == ModeSingleSchema || m.MigrationDefinition != previousMigrationDefinition {
			flattened = append(flattened, m.MigrationDefinition)
			previousMigrationDefinition = m.MigrationDefinition
		}
	}
	return flattened
}

func computeMigrationsToApply(diskMigrations []Migration, dbMigrations []DBMigration) []Migration {
	// flatten dbMigrations
	flattenedDBMigrations := flattenDBMigrations(dbMigrations)

	len := len(flattenedDBMigrations)
	var out []Migration

	// compute difference
	log.Println(len)
	for i := 0; i < len; i++ {
		log.Println(i)
		if diskMigrations[i].MigrationDefinition != flattenedDBMigrations[i] {
			out = append(out, diskMigrations[i])
		}
	}
	for _, m := range diskMigrations[len:] {
		out = append(out, m)
	}

	return out
}
