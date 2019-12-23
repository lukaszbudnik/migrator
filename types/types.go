package types

import (
	"time"
)

// MigrationType stores information about type of migration
type MigrationType uint32

const (
	// MigrationTypeSingleMigration is used to mark single migration
	MigrationTypeSingleMigration MigrationType = 1
	// MigrationTypeTenantMigration is used to mark tenant migrations
	MigrationTypeTenantMigration MigrationType = 2
	// MigrationTypeSingleScript is used to mark single SQL script which is executed always
	MigrationTypeSingleScript MigrationType = 3
	// MigrationTypeTenantScript is used to mark tenant SQL scripts which is executed always
	MigrationTypeTenantScript MigrationType = 4
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

// MigrationResults contains summary information about executed migrations
type MigrationResults struct {
	StartedAt             time.Time
	Duration              time.Duration
	Tenants               int
	SingleMigrations      int
	TenantMigrations      int
	TenantMigrationsTotal int // tenant migrations for all tenants
	MigrationsTotal       int // total number of all migrations applied
	SingleScripts         int
	TenantScripts         int
	TenantScriptsTotal    int // tenant scripts for all tenants
	ScriptsTotal          int // total number of all scripts applied
}
