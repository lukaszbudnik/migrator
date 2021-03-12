package types

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
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
type DBMigration struct {
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

// Summary contains summary information about executed migrations
type Summary struct {
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

// VersionInput is used by GraphQL to create new version in DB
type VersionInput struct {
	VersionName string
	Action      Action
	DryRun      bool
}

// TenantInput is used by GraphQL to create a new tenant in DB
type TenantInput struct {
	VersionName string
	Action      Action
	DryRun      bool
	TenantName  string
}

// APIVersion represents migrator API versions
type APIVersion string

const (
	// APIV1 - REST API - removed in v2021.0.0
	APIV1 APIVersion = "v1"
	// APIV2 - GraphQL API - current
	APIV2 APIVersion = "v2"
)

// VersionInfo contains build information and supported API versions
type VersionInfo struct {
	Release     string       `json:"release"`
	CommitSha   string       `json:"commitSha"`
	CommitDate  string       `json:"commitDate"`
	APIVersions []APIVersion `json:"apiVersions"`
}
