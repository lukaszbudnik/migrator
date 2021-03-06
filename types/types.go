package types

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
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

// ImplementsGraphQLType maps MigrationType Go type
// to the graphql scalar type in the schema
func (MigrationType) ImplementsGraphQLType(name string) bool {
	return name == "MigrationType"
}

// String converts MigrationType Go type to string literal
func (t MigrationType) String() string {
	switch t {
	case MigrationTypeSingleMigration:
		return "SingleMigration"
	case MigrationTypeTenantMigration:
		return "TenantMigration"
	case MigrationTypeSingleScript:
		return "SingleScript"
	case MigrationTypeTenantScript:
		return "TenantScript"
	default:
		panic(fmt.Sprintf("Unknown MigrationType value: %v", uint32(t)))
	}
}

// UnmarshalGraphQL converts string literal to MigrationType Go type
func (t *MigrationType) UnmarshalGraphQL(input interface{}) error {
	if str, ok := input.(string); ok {
		switch str {
		case "SingleMigration":
			*t = MigrationTypeSingleMigration
		case "TenantMigration":
			*t = MigrationTypeTenantMigration
		case "SingleScript":
			*t = MigrationTypeSingleScript
		case "TenantScript":
			*t = MigrationTypeTenantScript
		default:
			panic(fmt.Sprintf("Unknown MigrationType literal: %v", str))
		}
		return nil
	}
	return fmt.Errorf("Wrong type for MigrationType: %T", input)
}

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

// Tenant contains basic information about tenant
type Tenant struct {
	Name string `json:"name"`
}

// Version contains information about migrator versions
type Version struct {
	ID           int32         `json:"id"`
	Name         string        `json:"name"`
	Created      graphql.Time  `json:"created"`
	DBMigrations []DBMigration `json:"dbMigrations"`
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

// DBMigration embeds Migration and adds DB-specific fields
// replaces deprecated MigrationDB
type DBMigration = MigrationDB

// MigrationDB embeds Migration and adds DB-specific fields
// deprecated in v2020.1.0 sunset in v2021.1.0
// replaced by DBMigration
type MigrationDB struct {
	Migration
	ID     int32  `json:"id"`
	Schema string `json:"schema"`
	// appliedAt is deprecated the SQL column is already called created
	// API v1 uses AppliedAt
	// this field is ignored by GrapQL
	AppliedAt graphql.Time `json:"appliedAt"`
	// API v2 uses Created
	// this field is returned together with appliedAt
	// however it does not break API contract as this is a new field
	Created graphql.Time `json:"created"`
}

// Summary contains summary information about created version
// replaces deprecated MigrationDB
type Summary = MigrationResults

// MigrationResults contains summary information about executed migrations
// deprecated in v2020.1.0 sunset in v2021.1.0
// replaced by Stats
type MigrationResults struct {
	VersionID             int32        `json:"versionId"`
	StartedAt             graphql.Time `json:"startedAt"`
	Duration              int32        `json:"duration"`
	Tenants               int32        `json:"tenants"`
	SingleMigrations      int32        `json:"singleMigrations"`
	TenantMigrations      int32        `json:"tenantMigrations"`
	TenantMigrationsTotal int32        `json:"tenantMigrationsTotal"` // tenant migrations for all tenants
	MigrationsGrandTotal  int32        `json:"migrationsGrandTotal"`  // total number of all migrations applied
	SingleScripts         int32        `json:"singleScripts"`
	TenantScripts         int32        `json:"tenantScripts"`
	TenantScriptsTotal    int32        `json:"tenantScriptsTotal"` // tenant scripts for all tenants
	ScriptsGrandTotal     int32        `json:"scriptsGrandTotal"`  // total number of all scripts applied
}

// CreateResults contains results of CreateVersion or CreateTenant
type CreateResults struct {
	Summary *Summary
	Version *Version
}

// Action stores information about migrator action
type Action int

const (
	// ActionApply (the default action) tells migrator to apply all source migrations
	ActionApply Action = iota
	// ActionSync tells migrator to synchronise source migrations and not apply them
	ActionSync
)

// ImplementsGraphQLType maps Action Go type
// to the graphql scalar type in the schema
func (Action) ImplementsGraphQLType(name string) bool {
	return name == "Action"
}

// String converts MigrationType Go type to string literal
func (a Action) String() string {
	switch a {
	case ActionSync:
		return "Sync"
	case ActionApply:
		return "Apply"
	default:
		panic(fmt.Sprintf("Unknown Action value: %v", uint32(a)))
	}
}

// UnmarshalGraphQL converts string literal to MigrationType Go type
func (a *Action) UnmarshalGraphQL(input interface{}) error {
	if str, ok := input.(string); ok {
		switch str {
		case "Sync":
			*a = ActionSync
		case "Apply":
			*a = ActionApply
		default:
			panic(fmt.Sprintf("Unknown Action literal: %v", str))
		}
		return nil
	}
	return fmt.Errorf("Wrong type for Action: %T", input)
}

type VersionInput struct {
	VersionName string
	Action      Action
	DryRun      bool
}

type TenantInput struct {
	VersionName string
	Action      Action
	DryRun      bool
	TenantName  string
}

// VersionInfo contains build information and supported API versions
type VersionInfo struct {
	Release     string   `json:"release"`
	CommitSha   string   `json:"commitSha"`
	CommitDate  string   `json:"commitDate"`
	APIVersions []string `json:"apiVersions"`
}
