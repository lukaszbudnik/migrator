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

// Migration contains basic information about migration
type Migration struct {
	Name          string
	SourceDir     string
	File          string
	MigrationType MigrationType
	Contents      string
	CheckSum      string
}

// MigrationDB embeds Migration and adds DB-specific fields
type MigrationDB struct {
	Migration
	Schema  string
	Created time.Time
}
