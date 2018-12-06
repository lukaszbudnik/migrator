package core

import (
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/migrations"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/lukaszbudnik/migrator/utils"
	"log"
)

const (
	// DefaultConfigFile is a default migrator configuration file name
	DefaultConfigFile = "migrator.yaml"
	// ApplyAction is an action which applies disk migrations to database
	ApplyAction = "apply"
	// AddTenantAction is an action which creates new tenant in database
	// and applies all known migrations
	AddTenantAction = "addTenant"
	// PrintConfigAction is an action which prints contents of config
	PrintConfigAction = "config"
	// GetDBMigrationsAction is an action which lists migrations recorded in DB
	GetDBMigrationsAction = "dbMigrations"
	// GetDBTenantsAction is an action which list tenants for multi-tenant schemas
	GetDBTenantsAction = "dbTenants"
	// GetDiskMigrationsAction is an action which lists migrations stored on disk
	GetDiskMigrationsAction = "diskMigrations"
	// ServerMode is a mode for using migrator HTTP server
	ServerMode = "server"
	// ToolMode is a mode for using migrator command line
	ToolMode = "tool"
)

// ExecuteFlags is used to group flags passed to migrator when running in tool mode
type ExecuteFlags struct {
	Action string
	Tenant string
}

// GetDiskMigrations is a function which loads all migrations from disk as defined in config passed as first argument
// and using loader created by a function passed as second argument
func GetDiskMigrations(config *config.Config, createLoader func(*config.Config) loader.Loader) []types.Migration {
	loader := createLoader(config)
	diskMigrations := loader.GetDiskMigrations()
	log.Printf("Read disk migrations: %d", len(diskMigrations))
	return diskMigrations
}

// GetDBTenants is a function which loads all tenants for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func GetDBTenants(config *config.Config, createConnector func(*config.Config) db.Connector) []string {
	connector := createConnector(config)
	connector.Init()
	defer connector.Dispose()
	dbTenants := connector.GetTenants()
	log.Printf("Read DB tenants: %d", len(dbTenants))
	return dbTenants
}

// GetDBMigrations is a function which loads all DB migrations for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func GetDBMigrations(config *config.Config, createConnector func(*config.Config) db.Connector) []types.MigrationDB {
	connector := createConnector(config)
	connector.Init()
	defer connector.Dispose()
	dbMigrations := connector.GetDBMigrations()
	log.Printf("Read DB migrations: %d", len(dbMigrations))
	return dbMigrations
}

// ApplyMigrations is a function which applies disk migrations to DB as defined in config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func ApplyMigrations(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) []types.Migration {
	diskMigrations := GetDiskMigrations(config, createLoader)
	dbMigrations := GetDBMigrations(config, createConnector)
	migrationsToApply := migrations.ComputeMigrationsToApply(diskMigrations, dbMigrations)

	log.Printf("Found migrations to apply: %d", len(migrationsToApply))
	doApplyMigrations(migrationsToApply, config, createConnector)

	notifier := notifications.CreateNotifier(config)
	text := fmt.Sprintf("Migrations applied: %d", len(migrationsToApply))
	resp, err := notifier.Notify(text)

	if err != nil {
		log.Printf("Notifier err: %v", err)
	} else {
		log.Printf("Notifier response: %v", resp)
	}

	return migrationsToApply
}

// AddTenant creates new tenant in DB and applies all tenant migrations
func AddTenant(tenant string, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) []types.Migration {

	diskMigrations := GetDiskMigrations(config, createLoader)

	// filter only tenant schemas
	// var migrationsToApply []types.Migration
	migrationsToApply := migrations.FilterTenantMigrations(diskMigrations)

	log.Printf("Found migrations to apply: %d", len(migrationsToApply))
	doAddTenantAndApplyMigrations(tenant, migrationsToApply, config, createConnector)

	notifier := notifications.CreateNotifier(config)
	text := fmt.Sprintf("Tenant %q added, migrations applied: %d", tenant, len(migrationsToApply))
	resp, err := notifier.Notify(text)

	if err != nil {
		log.Printf("Notifier err: %v", err)
	} else {
		log.Printf("Notifier response: %v", resp)
	}

	return diskMigrations
}

// ExecuteMigrator is a function which executes actions on resources defined in config passed as first argument action defined as second argument
// and using connector created by a function passed as third argument and disk loader created by a function passed as fourth argument
func ExecuteMigrator(config *config.Config, executeFlags ExecuteFlags) {
	doExecuteMigrator(config, executeFlags, db.CreateConnector, loader.CreateLoader)
}

func doExecuteMigrator(config *config.Config, executeFlags ExecuteFlags, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	switch executeFlags.Action {
	case PrintConfigAction:
		log.Printf("Configuration file ==>\n%v\n", config)
	case GetDiskMigrationsAction:
		diskMigrations := GetDiskMigrations(config, createLoader)
		if len(diskMigrations) > 0 {
			log.Printf("List of disk migrations\n%v", utils.MigrationArrayToString(diskMigrations))
		}
	case GetDBMigrationsAction:
		dbMigrations := GetDBMigrations(config, createConnector)
		if len(dbMigrations) > 0 {
			log.Printf("List of db migrations\n%v", utils.MigrationDBArrayToString(dbMigrations))
		}
	case AddTenantAction:
		AddTenant(executeFlags.Tenant, config, createConnector, createLoader)
	case GetDBTenantsAction:
		dbTenants := GetDBTenants(config, createConnector)
		if len(dbTenants) > 0 {
			log.Printf("List of db tenants\n%v", utils.TenantArrayToString(dbTenants))
		}
	case ApplyAction:
		migrationsApplied := ApplyMigrations(config, createConnector, createLoader)
		if len(migrationsApplied) > 0 {
			log.Printf("List of migrations applied\n%v", utils.MigrationArrayToString(migrationsApplied))
		}
	}
}

func doApplyMigrations(migrationsToApply []types.Migration, config *config.Config, createConnector func(*config.Config) db.Connector) {
	connector := createConnector(config)
	connector.Init()
	defer connector.Dispose()
	connector.ApplyMigrations(migrationsToApply)
}

func doAddTenantAndApplyMigrations(tenant string, migrationsToApply []types.Migration, config *config.Config, createConnector func(*config.Config) db.Connector) {
	connector := createConnector(config)
	connector.Init()
	defer connector.Dispose()
	connector.AddTenantAndApplyMigrations(tenant, migrationsToApply)
}
