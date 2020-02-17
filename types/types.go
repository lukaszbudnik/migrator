package types

import (
	"encoding/json"
	"fmt"
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

// ImplementsGraphQLType maps this custom Go type
// to the graphql scalar type in the schema.
func (MigrationType) ImplementsGraphQLType(name string) bool {
	return name == "MigrationType"
}

// This function will be called whenever you use the
// time scalar as an input
func (t *MigrationType) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case MigrationType:
		t = &input
		return nil
	case uint32:
		mt := MigrationType(input)
		t = &mt
		return nil
	case int:
		fmt.Printf("I'm int %v\n", input)
		mt := MigrationType(input)
		fmt.Printf("I'm MT %v\n", mt)
		fmt.Printf("I'm MT %v\n", &mt)
		t = &mt
		return nil
	default:
		return fmt.Errorf("wrong type %v", input)
	}
}

// This function will be called whenever you
// query for fields that use the Time type
func (t MigrationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(t))
}

type MigrationTypeOptional struct {
	MigrationType MigrationType
}

// ImplementsGraphQLType maps this custom Go type
// to the graphql scalar type in the schema.
func (MigrationTypeOptional) ImplementsGraphQLType(name string) bool {
	return name == "MigrationTypeOptional"
}

// This function will be called whenever you use the
// time scalar as an input
func (t *MigrationTypeOptional) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case MigrationType:
		t.MigrationType = input
		return nil
	case uint32:
		mt := MigrationType(input)
		t.MigrationType = mt
		return nil
	case int:
		mt := MigrationType(input)
		t.MigrationType = mt
		return nil
	default:
		return fmt.Errorf("Could not unmarshal graphql MigrationTypeOptional: %v", input)
	}
}

// This function will be called whenever you
// query for fields that use the Time type
func (t MigrationTypeOptional) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(t.MigrationType))
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
