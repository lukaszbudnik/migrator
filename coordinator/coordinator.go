package coordinator

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
)

// Coordinator interface abstracts all operations performed by migrator
type Coordinator interface {
	GetTenants() []types.Tenant
	GetVersions() []types.Version
	GetVersionsByFile(string) []types.Version
	GetSourceMigrations() []types.Migration
	GetAppliedMigrations() []types.MigrationDB
	VerifySourceMigrationsCheckSums() (bool, []types.Migration)
	ApplyMigrations(types.MigrationsModeType) (*types.MigrationResults, []types.Migration)
	AddTenantAndApplyMigrations(types.MigrationsModeType, string) (*types.MigrationResults, []types.Migration)
	Dispose()
}

// coordinator struct is a struct for implementing DB specific dialects
type coordinator struct {
	ctx               context.Context
	connector         db.Connector
	loader            loader.Loader
	notifier          notifications.Notifier
	config            *config.Config
	tenants           []types.Tenant
	versions          []types.Version
	sourceMigrations  []types.Migration
	appliedMigrations []types.MigrationDB
	loaderLock        sync.Mutex
	connectorLock     sync.Mutex
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
	c.connectorLock.Lock()
	defer c.connectorLock.Unlock()
	if c.tenants == nil {
		tenants := c.connector.GetTenants()
		c.tenants = tenants
	}
	return c.tenants
}

func (c *coordinator) GetVersions() []types.Version {
	c.connectorLock.Lock()
	defer c.connectorLock.Unlock()
	if c.versions == nil {
		versions := c.connector.GetVersions()
		c.versions = versions
	}
	return c.versions
}

func (c *coordinator) GetVersionsByFile(file string) []types.Version {
	return c.connector.GetVersionsByFile(file)
}

func (c *coordinator) GetSourceMigrations() []types.Migration {
	c.loaderLock.Lock()
	defer c.loaderLock.Unlock()
	if c.sourceMigrations == nil {
		sourceMigrations := c.loader.GetSourceMigrations()
		c.sourceMigrations = sourceMigrations
	}
	return c.sourceMigrations
}

func (c *coordinator) GetAppliedMigrations() []types.MigrationDB {
	c.connectorLock.Lock()
	defer c.connectorLock.Unlock()
	if c.appliedMigrations == nil {
		appliedMigrations := c.connector.GetAppliedMigrations()
		c.appliedMigrations = appliedMigrations
	}
	return c.appliedMigrations
}

// VerifySourceMigrationsCheckSums verifies if CheckSum of source and applied DB migrations match
// VerifySourceMigrationsCheckSums allows CheckSum of scripts to be different (they are applied every time and are often updated)
// returns bool indicating if offending (i.e., modified) disk migrations were found
// if bool is false the function returns a slice of offending migrations
// if bool is true the slice of effending migrations is empty
func (c *coordinator) VerifySourceMigrationsCheckSums() (bool, []types.Migration) {
	sourceMigrations := c.GetSourceMigrations()
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
	sourceMigrations := c.GetSourceMigrations()
	appliedMigrations := c.GetAppliedMigrations()

	migrationsToApply := c.computeMigrationsToApply(sourceMigrations, appliedMigrations)
	common.LogInfo(c.ctx, "Found migrations to apply: %d", len(migrationsToApply))

	results := c.connector.ApplyMigrations(mode, migrationsToApply)

	c.sendNotification(results)

	return results, migrationsToApply
}

func (c *coordinator) AddTenantAndApplyMigrations(mode types.MigrationsModeType, tenant string) (*types.MigrationResults, []types.Migration) {
	sourceMigrations := c.GetSourceMigrations()

	// filter only tenant schemas
	migrationsToApply := c.filterTenantMigrations(sourceMigrations)
	common.LogInfo(c.ctx, "Migrations to apply for new tenant: %d", len(migrationsToApply))

	results := c.connector.AddTenantAndApplyMigrations(mode, tenant, migrationsToApply)

	c.sendNotification(results)

	return results, migrationsToApply
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
