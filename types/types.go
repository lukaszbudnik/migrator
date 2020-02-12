package types

import (
	"time"

	"gopkg.in/go-playground/validator.v9"
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

// MigrationsResponseType represents type of response either full or summary
type MigrationsResponseType string

const (
	// ResponseTypeSummary instructs migrator to only return JSON representation of Results struct
	ResponseTypeSummary MigrationsResponseType = "summary"
	// ResponseTypeFull instructs migrator to return JSON representation of both Results struct and all applied migrations/scripts
	ResponseTypeFull MigrationsResponseType = "full"
	// ResponseTypeList instructs migrator to return JSON representation of both Results struct and all applied migrations/scripts but without their contents
	ResponseTypeList MigrationsResponseType = "list"
)

// MigrationsModeType represents mode in which migrations should be applied
type MigrationsModeType string

const (
	// ModeTypeApply instructs migrator to apply migrations
	ModeTypeApply MigrationsModeType = "apply"
	// ModeTypeDryRun instructs migrator to perform apply operation in dry-run mode, instead of committing transaction it is rollbacked
	ModeTypeDryRun MigrationsModeType = "dry-run"
	// ModeTypeSync instructs migrator to only synchronise migrations
	ModeTypeSync MigrationsModeType = "sync"
)

// ValidateMigrationsModeType validates MigrationsModeType used by binding package
func ValidateMigrationsModeType(fl validator.FieldLevel) bool {
	mode, ok := fl.Field().Interface().(MigrationsModeType)
	if ok {
		return mode == ModeTypeApply || mode == ModeTypeSync || mode == ModeTypeDryRun
	}
	return false
}

// ValidateMigrationsResponseType validates MigrationsResponseType used by binding package
func ValidateMigrationsResponseType(fl validator.FieldLevel) bool {
	response, ok := fl.Field().Interface().(MigrationsResponseType)
	if ok {
		return response == ResponseTypeSummary || response == ResponseTypeFull || response == ResponseTypeList
	}
	return false
}

// Migration contains basic information about migration
type Migration struct {
	Name          string        `json:"name"`
	SourceDir     string        `json:"sourceDir"`
	File          string        `json:"file"`
	MigrationType MigrationType `json:"migrationType"`
	Contents      string        `json:"contents,omitempty"`
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

// VersionInfo contains build information and supported API versions
type VersionInfo struct {
	Release     string   `json:"release"`
	CommitSha   string   `json:"commitSha"`
	CommitDate  string   `json:"commitDate"`
	APIVersions []string `json:"apiVersions"`
}
