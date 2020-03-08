package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
)

// SourceMigrationFilters defines filters which can be used to fetch source migrations
type SourceMigrationFilters struct {
	Name          *string
	SourceDir     *string
	File          *string
	MigrationType *types.MigrationType
}

// Coordinator interface abstracts all operations performed by migrator
type Coordinator interface {
	GetTenants() []types.Tenant
	GetVersions() []types.Version
	GetVersionsByFile(string) []types.Version
	GetVersionByID(int32) (*types.Version, error)
	GetDBMigrationByID(int32) (*types.DBMigration, error)
	GetSourceMigrations(*SourceMigrationFilters) []types.Migration
	GetSourceMigrationByFile(string) (*types.Migration, error)
	// deprecated in v2020.1.0 sunset in v2021.1.0
	// Version now contains slice of DBMigration
	GetAppliedMigrations() []types.MigrationDB
	VerifySourceMigrationsCheckSums() (bool, []types.Migration)
	// Deprecated, uses CreateVersion under the hood
	ApplyMigrations(types.MigrationsModeType) (*types.MigrationResults, []types.Migration)
	// Deprecated, uses CreateTenant under the hood
	AddTenantAndApplyMigrations(types.MigrationsModeType, string) (*types.MigrationResults, []types.Migration)
	CreateVersion(string, types.Action, bool) *types.CreateResults
	CreateTenant(string, types.Action, bool, string) *types.CreateResults
	Dispose()
}

// coordinator struct is a struct for implementing DB specific dialects
type coordinator struct {
	ctx       context.Context
	connector db.Connector
	loader    loader.Loader
	notifier  notifications.Notifier
	config    *config.Config
}

// Factory creates new Coordinator instance
type Factory func(context.Context, *config.Config) Coordinator

// New creates instance of Coordinator
func New(ctx context.Context, config *config.Config, newConnector db.Factory, newLoader loader.Factory, newNotifier notifications.Factory) Coordinator {
	connector := newConnector(ctx, config)
	loader := newLoader(ctx, config)
	notifier := newNotifier(ctx, config)
	coordinator := &coordinator{
		connector: connector,
		loader:    loader,
		notifier:  notifier,
		config:    config,
		ctx:       ctx,
	}
	return coordinator
}

func (c *coordinator) GetTenants() []types.Tenant {
	return c.connector.GetTenants()
}

func (c *coordinator) GetVersions() []types.Version {
	return c.connector.GetVersions()
}

func (c *coordinator) GetVersionsByFile(file string) []types.Version {
	return c.connector.GetVersionsByFile(file)
}

func (c *coordinator) GetVersionByID(ID int32) (*types.Version, error) {
	return c.connector.GetVersionByID(ID)
}

func (c *coordinator) GetSourceMigrations(filters *SourceMigrationFilters) []types.Migration {
	allSourceMigrations := c.loader.GetSourceMigrations()
	filteredMigrations := c.filterMigrations(allSourceMigrations, filters)
	return filteredMigrations
}

func (c *coordinator) GetSourceMigrationByFile(file string) (*types.Migration, error) {
	allSourceMigrations := c.loader.GetSourceMigrations()
	filters := SourceMigrationFilters{
		File: &file,
	}
	filteredMigrations := c.filterMigrations(allSourceMigrations, &filters)
	if len(filteredMigrations) == 0 {
		return nil, fmt.Errorf("Source migration not found: %v", file)
	}
	return &filteredMigrations[0], nil
}

func (c *coordinator) GetDBMigrationByID(ID int32) (*types.DBMigration, error) {
	return c.connector.GetDBMigrationByID(ID)
}

func (c *coordinator) GetAppliedMigrations() []types.MigrationDB {
	return c.connector.GetAppliedMigrations()
}

// VerifySourceMigrationsCheckSums verifies if CheckSum of source and applied DB migrations match
// VerifySourceMigrationsCheckSums allows CheckSum of scripts to be different (they are applied every time and are often updated)
// returns bool indicating if offending (i.e., modified) disk migrations were found
// if bool is false the function returns a slice of offending migrations
// if bool is true the slice of effending migrations is empty
func (c *coordinator) VerifySourceMigrationsCheckSums() (bool, []types.Migration) {
	sourceMigrations := c.GetSourceMigrations(nil)
	appliedMigrations := c.GetAppliedMigrations()

	flattenedAppliedMigration := c.flattenAppliedMigrations(appliedMigrations)

	intersect := c.intersect(sourceMigrations, flattenedAppliedMigration)

	var offendingMigrations []types.Migration
	var result = true
	for _, t := range intersect {
		if t.source.MigrationType == types.MigrationTypeSingleScript || t.source.MigrationType == types.MigrationTypeTenantScript {
			continue
		}
		if t.source.CheckSum != t.applied.CheckSum {
			offendingMigrations = append(offendingMigrations, t.source)
			result = false
		}
	}
	return result, offendingMigrations
}

func (c *coordinator) ApplyMigrations(mode types.MigrationsModeType) (*types.MigrationResults, []types.Migration) {

	// convert to new API params
	versionName := "createVersion"
	action := types.ActionApply
	dryRun := false
	if mode == types.ModeTypeDryRun {
		dryRun = true
	} else if mode == types.ModeTypeSync {
		action = types.ActionSync
	}

	sourceMigrations := c.GetSourceMigrations(nil)
	appliedMigrations := c.GetAppliedMigrations()

	migrationsToApply := c.computeMigrationsToApply(sourceMigrations, appliedMigrations)
	common.LogInfo(c.ctx, "Found migrations to apply: %d", len(migrationsToApply))

	results, _ := c.connector.CreateVersion(versionName, action, dryRun, migrationsToApply)

	c.sendNotification(results)

	return results, migrationsToApply
}

func (c *coordinator) CreateVersion(versionName string, action types.Action, dryRun bool) *types.CreateResults {
	sourceMigrations := c.GetSourceMigrations(nil)
	appliedMigrations := c.GetAppliedMigrations()

	migrationsToApply := c.computeMigrationsToApply(sourceMigrations, appliedMigrations)
	common.LogInfo(c.ctx, "Found migrations to apply: %d", len(migrationsToApply))

	summary, version := c.connector.CreateVersion(versionName, action, dryRun, migrationsToApply)

	c.sendNotification(summary)

	return &types.CreateResults{Summary: summary, Version: version}
}

func (c *coordinator) AddTenantAndApplyMigrations(mode types.MigrationsModeType, tenant string) (*types.MigrationResults, []types.Migration) {
	// convert to new API params
	versionName := "createVersion"
	action := types.ActionApply
	dryRun := false
	if mode == types.ModeTypeDryRun {
		dryRun = true
	} else if mode == types.ModeTypeSync {
		action = types.ActionSync
	}

	sourceMigrations := c.GetSourceMigrations(nil)

	// filter only tenant schemas
	migrationsToApply := c.filterTenantMigrations(sourceMigrations)
	common.LogInfo(c.ctx, "Migrations to apply for new tenant: %d", len(migrationsToApply))

	summary, _ := c.connector.CreateTenant(versionName, action, dryRun, tenant, migrationsToApply)

	c.sendNotification(summary)

	return summary, migrationsToApply
}

func (c *coordinator) CreateTenant(versionName string, action types.Action, dryRun bool, tenant string) *types.CreateResults {
	sourceMigrations := c.GetSourceMigrations(nil)

	// filter only tenant schemas
	migrationsToApply := c.filterTenantMigrations(sourceMigrations)
	common.LogInfo(c.ctx, "Migrations to apply for new tenant: %d", len(migrationsToApply))

	summary, version := c.connector.CreateTenant(versionName, action, dryRun, tenant, migrationsToApply)

	c.sendNotification(summary)

	return &types.CreateResults{Summary: summary, Version: version}
}

func (c *coordinator) Dispose() {
	c.connector.Dispose()
}

func (c *coordinator) flattenAppliedMigrations(appliedMigrations []types.MigrationDB) []types.Migration {
	var flattened []types.Migration
	var previousMigration types.Migration
	for i, m := range appliedMigrations {
		if i == 0 || m.Migration != previousMigration {
			flattened = append(flattened, m.Migration)
			previousMigration = m.Migration
		}
	}
	return flattened
}

// intersect returns the elements from source and applied
func (c *coordinator) intersect(sourceMigrations []types.Migration, flattenedAppliedMigrations []types.Migration) []struct {
	source  types.Migration
	applied types.Migration
} {
	// key is Migration.File
	existsInDB := map[string]types.Migration{}
	for _, m := range flattenedAppliedMigrations {
		existsInDB[m.File] = m
	}
	intersect := []struct {
		source  types.Migration
		applied types.Migration
	}{}
	for _, m := range sourceMigrations {
		if db, ok := existsInDB[m.File]; ok {
			intersect = append(intersect, struct {
				source  types.Migration
				applied types.Migration
			}{m, db})
		}
	}
	return intersect
}

// difference returns the elements on disk which are not yet in DB
// the exceptions are MigrationTypeSingleScript and MigrationTypeTenantScript which are always run
func (c *coordinator) difference(sourceMigrations []types.Migration, flattenedAppliedMigrations []types.Migration) []types.Migration {
	// key is Migration.File
	existsInDB := map[string]bool{}
	for _, m := range flattenedAppliedMigrations {
		if m.MigrationType != types.MigrationTypeSingleScript && m.MigrationType != types.MigrationTypeTenantScript {
			existsInDB[m.File] = true
		}
	}
	diff := []types.Migration{}
	for _, m := range sourceMigrations {
		if _, ok := existsInDB[m.File]; !ok {
			diff = append(diff, m)
		}
	}
	return diff
}

// computeMigrationsToApply computes which source migrations should be applied to DB based on migrations already present in DB
func (c *coordinator) computeMigrationsToApply(sourceMigrations []types.Migration, appliedMigrations []types.MigrationDB) []types.Migration {
	flattenedAppliedMigrations := c.flattenAppliedMigrations(appliedMigrations)

	len := len(flattenedAppliedMigrations)
	common.LogInfo(c.ctx, "Number of flattened DB migrations: %d", len)

	out := c.difference(sourceMigrations, flattenedAppliedMigrations)
	return out
}

// filterTenantMigrations returns only migrations which are of type MigrationTypeTenantSchema
func (c *coordinator) filterTenantMigrations(sourceMigrations []types.Migration) []types.Migration {
	filteredTenantMigrations := []types.Migration{}
	for _, m := range sourceMigrations {
		if m.MigrationType == types.MigrationTypeTenantMigration || m.MigrationType == types.MigrationTypeTenantScript {
			filteredTenantMigrations = append(filteredTenantMigrations, m)
		}
	}

	return filteredTenantMigrations
}

// errors are silently discarded, adding tenant or applying migrations
// must not fail because of notification error
func (c *coordinator) sendNotification(results *types.MigrationResults) {
	bytes, _ := json.Marshal(results)
	text := string(bytes)
	if resp, err := c.notifier.Notify(text); err != nil {
		common.LogError(c.ctx, "Notifier error: %v", err.Error())
	} else {
		common.LogInfo(c.ctx, "Notifier response: %v", resp)
	}
}

func (c *coordinator) filterMigrations(migrations []types.Migration, filters *SourceMigrationFilters) []types.Migration {
	filtered := []types.Migration{}
	for _, m := range migrations {
		match := c.matchMigration(m, filters)
		if match {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func (c *coordinator) matchMigration(m types.Migration, filters *SourceMigrationFilters) bool {
	match := true

	if filters == nil {
		return match
	}

	ps := reflect.ValueOf(filters)
	s := ps.Elem()
	for i := 0; i < s.Type().NumField(); i++ {
		// if filter is nil it means match
		if s.Field(i).IsNil() {
			continue
		}
		// args field names match migration names
		pm := reflect.ValueOf(m).FieldByName(s.Type().Field(i).Name)
		match = match && (pm.Interface() == s.Field(i).Elem().Interface())
		// if already non match don't bother further looping
		if !match {
			break
		}
	}
	return match
}
