package core

import (
	"errors"
	"fmt"
	"log"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/migrations"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/lukaszbudnik/migrator/utils"
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
func GetDiskMigrations(config *config.Config, createLoader func(*config.Config) loader.Loader) ([]types.Migration, error) {
	loader := createLoader(config)
	return loader.GetDiskMigrations()
}

// GetDBTenants is a function which loads all tenants for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func GetDBTenants(config *config.Config, createConnector func(*config.Config) db.Connector) ([]string, error) {
	connector := createConnector(config)
	if err := connector.Init(); err != nil {
		return []string{}, err
	}
	defer connector.Dispose()
	return connector.GetTenants()
}

// GetDBMigrations is a function which loads all DB migrations for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func GetDBMigrations(config *config.Config, createConnector func(*config.Config) db.Connector) ([]types.MigrationDB, error) {
	connector := createConnector(config)
	if err := connector.Init(); err != nil {
		return []types.MigrationDB{}, err
	}
	defer connector.Dispose()
	return connector.GetDBMigrations()
}

// ApplyMigrations is a function which applies disk migrations to DB as defined in config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func ApplyMigrations(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) (migrationsToApply []types.Migration, err error) {
	diskMigrations, err := GetDiskMigrations(config, createLoader)
	if err != nil {
		return
	}
	log.Printf("Read disk migrations: %d", len(diskMigrations))

	dbMigrations, err := GetDBMigrations(config, createConnector)
	if err != nil {
		return
	}
	log.Printf("Read DB migrations: %d", len(dbMigrations))

	migrationsToApply = migrations.ComputeMigrationsToApply(diskMigrations, dbMigrations)
	log.Printf("Found migrations to apply: %d", len(migrationsToApply))

	err = doApplyMigrations(migrationsToApply, config, createConnector)
	if err != nil {
		return
	}

	text := fmt.Sprintf("Migrations applied: %d", len(migrationsToApply))
	sendNotification(config, text)

	return
}

// AddTenant creates new tenant in DB and applies all tenant migrations
func AddTenant(tenant string, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) (migrationsToApply []types.Migration, err error) {

	diskMigrations, err := GetDiskMigrations(config, createLoader)
	if err != nil {
		return
	}
	log.Printf("Read disk migrations: %d", len(diskMigrations))

	// filter only tenant schemas
	migrationsToApply = migrations.FilterTenantMigrations(diskMigrations)
	log.Printf("Found migrations to apply: %d", len(migrationsToApply))

	err = doAddTenantAndApplyMigrations(tenant, migrationsToApply, config, createConnector)
	if err != nil {
		return
	}

	text := fmt.Sprintf("Tenant %q added, migrations applied: %d", tenant, len(migrationsToApply))
	sendNotification(config, text)

	return
}

// VerifyMigrations loads disk and db migrations and verifies their checksums
// see migrations.VerifyCheckSums for more information
func VerifyMigrations(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) (bool, []types.Migration, error) {
	diskMigrations, err := GetDiskMigrations(config, createLoader)
	if err != nil {
		return false, []types.Migration{}, err
	}

	dbMigrations, err := GetDBMigrations(config, createConnector)
	if err != nil {
		return false, []types.Migration{}, err
	}
	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)
	return verified, offendingMigrations, nil
}

// ExecuteMigrator is a function which executes actions on resources defined in config passed as first argument action defined as second argument
// and using connector created by a function passed as third argument and disk loader created by a function passed as fourth argument
func ExecuteMigrator(config *config.Config, executeFlags ExecuteFlags) {
	err := doExecuteMigrator(config, executeFlags, db.CreateConnector, loader.CreateLoader)
	if err != nil {
		log.Printf("Error encountered: %v", err)
	}
}

func doExecuteMigrator(config *config.Config, executeFlags ExecuteFlags, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) error {
	switch executeFlags.Action {
	case PrintConfigAction:
		log.Printf("Configuration file ==>\n%v\n", config)
	case GetDiskMigrationsAction:
		diskMigrations, err := GetDiskMigrations(config, createLoader)
		if err != nil {
			return err
		}
		if len(diskMigrations) > 0 {
			log.Printf("List of disk migrations\n%v", utils.MigrationArrayToString(diskMigrations))
		}
	case GetDBMigrationsAction:
		dbMigrations, err := GetDBMigrations(config, createConnector)
		if err != nil {
			return err
		}
		log.Printf("Read DB migrations: %d", len(dbMigrations))
		if len(dbMigrations) > 0 {
			log.Printf("List of db migrations\n%v", utils.MigrationDBArrayToString(dbMigrations))
		}
	case AddTenantAction:
		verified, offendingMigrations, err := VerifyMigrations(config, createConnector, createLoader)
		if err != nil {
			return err
		}
		if !verified {
			log.Printf("Checksum verification failed.")
			log.Printf("List of offending disk migrations\n%v", utils.MigrationArrayToString(offendingMigrations))
			return errors.New("Checksum verification failed")
		}

		migrationsApplied, err := AddTenant(executeFlags.Tenant, config, createConnector, createLoader)
		if err != nil {
			return err
		}
		if len(migrationsApplied) > 0 {
			log.Printf("List of migrations applied\n%v", utils.MigrationArrayToString(migrationsApplied))
		}
	case GetDBTenantsAction:
		dbTenants, err := GetDBTenants(config, createConnector)
		if err != nil {
			return err
		}
		log.Printf("Read DB tenants: %d", len(dbTenants))
		if len(dbTenants) > 0 {
			log.Printf("List of db tenants\n%v", utils.TenantArrayToString(dbTenants))
		}
	case ApplyAction:
		verified, offendingMigrations, err := VerifyMigrations(config, createConnector, createLoader)
		if err != nil {
			return err
		}
		if !verified {
			log.Printf("Checksum verification failed.")
			log.Printf("List of offending disk migrations\n%v", utils.MigrationArrayToString(offendingMigrations))
			return errors.New("Checksum verification failed")
		}
		migrationsApplied, err := ApplyMigrations(config, createConnector, createLoader)
		if err != nil {
			return err
		}
		if len(migrationsApplied) > 0 {
			log.Printf("List of migrations applied\n%v", utils.MigrationArrayToString(migrationsApplied))
		}
	}
	return nil
}

func doApplyMigrations(migrationsToApply []types.Migration, config *config.Config, createConnector func(*config.Config) db.Connector) error {
	connector := createConnector(config)
	if err := connector.Init(); err != nil {
		return err
	}
	defer connector.Dispose()
	return connector.ApplyMigrations(migrationsToApply)
}

func doAddTenantAndApplyMigrations(tenant string, migrationsToApply []types.Migration, config *config.Config, createConnector func(*config.Config) db.Connector) error {
	connector := createConnector(config)
	if err := connector.Init(); err != nil {
		return err
	}
	defer connector.Dispose()
	return connector.AddTenantAndApplyMigrations(tenant, migrationsToApply)
}

func sendNotification(config *config.Config, text string) {
	notifier := notifications.CreateNotifier(config)
	resp, err := notifier.Notify(text)

	if err != nil {
		log.Printf("Notifier err: %v", err)
	} else {
		log.Printf("Notifier response: %v", resp)
	}
}
