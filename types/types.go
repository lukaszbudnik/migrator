package types

import (
	"time"
)

// MigrationType stores information about type of migration
type MigrationType uint32

const (
	// MigrationTypeSingleSchema is used to mark single migration
	MigrationTypeSingleSchema MigrationType = 1
	// MigrationTypeTenantSchema is used to mark tenant migrations
	MigrationTypeTenantSchema MigrationType = 2
)

// MigrationDefinition contains basic information about migration
type MigrationDefinition struct {
	Name          string
	SourceDir     string
	File          string
	MigrationType MigrationType
}

// Migration embeds MigrationDefinition and contains its contents
type Migration struct {
	MigrationDefinition
	Contents string
}

// MigrationDB embeds MigrationDefinition and contain other DB properties
type MigrationDB struct {
	MigrationDefinition
	Schema  string
	Created time.Time
}
