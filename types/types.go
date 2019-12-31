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

type MigrationsResponseType string

const (
	ResponseTypeSummary MigrationsResponseType = "summary"
	ResponseTypeFull    MigrationsResponseType = "full"
)

type MigrationsModeType string

const (
	ModeTypeApply  MigrationsModeType = "apply"
	ModeTypeSync   MigrationsModeType = "sync"
	ModeTypeDryRun MigrationsModeType = "dry-run"
)

func ValidateMigrationsMode(mode MigrationsModeType) bool {
	return mode == ModeTypeApply || mode == ModeTypeSync || mode == ModeTypeDryRun
}

// Migration contains basic information about migration
type Migration struct {
	Name          string        `json:"name"`
	SourceDir     string        `json:"sourceDir"`
	File          string        `json:"file"`
	MigrationType MigrationType `json:"migrationType"`
	Contents      string        `json:"contents"`
	CheckSum      string        `json:"checkSum"`
}

// MigrationDB embeds Migration and adds DB-specific fields
type MigrationDB struct {
	Migration
	Schema    string    `json:"schema"`
	AppliedAt time.Time `json:"appliedAt"`
}

// MigrationResults contains summary information about executed migrations
type MigrationResults struct {
	StartedAt             time.Time     `json:"startedAt"`
	Duration              time.Duration `json:"duration"`
	Tenants               int           `json:"tenants"`
	SingleMigrations      int           `json:"singleMigrations"`
	TenantMigrations      int           `json:"tenantMigrations"`
	TenantMigrationsTotal int           `json:"tenantMigrationsTotal"` // tenant migrations for all tenants
	MigrationsGrandTotal  int           `json:"migrationsGrandTotal"`  // total number of all migrations applied
	SingleScripts         int           `json:"singleScripts"`
	TenantScripts         int           `json:"tenantScripts"`
	TenantScriptsTotal    int           `json:"tenantScriptsTotal"` // tenant scripts for all tenants
	ScriptsGrandTotal     int           `json:"scriptsGrandTotal"`  // total number of all scripts applied
}
