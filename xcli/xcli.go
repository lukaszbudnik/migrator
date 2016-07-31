package xcli

import (
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/migrations"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
	"log"
)

const (
	// ApplyAction is an action which applies disk migrations to database
	ApplyAction = "apply"
	// PrintConfigAction is an action which prints contents of config
	PrintConfigAction = "config"
	// ListDBMigrationsAction is an action which lists migrations recorded in DB
	ListDBMigrationsAction = "dbMigrations"
	// ListDBTenantsAction is an action which list tenants for multi-tenant schemas
	ListDBTenantsAction = "dbTenants"
	// ListDiskMigrationsAction is an action which lists migrations stored on disk
	ListDiskMigrationsAction = "diskMigrations"
)

// ReadConfig is a function which reads config from a file which names is passed as argument
func ReadConfig(configFile *string) *config.Config {
	config := config.FromFile(*configFile)
	log.Printf("Read config file ==> OK")
	return config
}

// LoadDiskMigrations is a function which loads all migrations from disk as defined in config passed as first argument
// and using loader created by a function passed as second argument
func LoadDiskMigrations(config *config.Config, createLoader func(*config.Config) loader.Loader) []types.Migration {
	loader := createLoader(config)
	diskMigrations := loader.GetMigrations()
	log.Printf("Read [%d] disk migrations ==> OK", len(diskMigrations))
	return diskMigrations
}

// LoadDBTenants is a function which loads all tenants for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func LoadDBTenants(config *config.Config, createConnector func(*config.Config) db.Connector) []string {
	connector := createConnector(config)
	connector.Init()
	defer connector.Dispose()
	dbTenants := connector.GetTenants()
	log.Printf("Read [%d] DB tenants ==> OK", len(dbTenants))
	return dbTenants
}

// LoadDBMigrations is a function which loads all DB migrations for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func LoadDBMigrations(config *config.Config, createConnector func(*config.Config) db.Connector) []types.MigrationDB {
	connector := createConnector(config)
	connector.Init()
	defer connector.Dispose()
	dbMigrations := connector.GetMigrations()
	log.Printf("Read [%d] DB migrations ==> OK", len(dbMigrations))
	return dbMigrations
}

func doApplyMigrations(migrationsToApply []types.Migration, config *config.Config, createConnector func(*config.Config) db.Connector) {
	connector := createConnector(config)
	connector.Init()
	defer connector.Dispose()
	connector.ApplyMigrations(migrationsToApply)
}

// ApplyMigrations is a function which applies disk migrations to DB as defined in config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func ApplyMigrations(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) []types.Migration {
	diskMigrations := LoadDiskMigrations(config, createLoader)
	dbMigrations := LoadDBMigrations(config, createConnector)
	migrationsToApply := migrations.ComputeMigrationsToApply(diskMigrations, dbMigrations)

	log.Printf("Found [%d] migrations to apply ==> OK", len(migrationsToApply))
	doApplyMigrations(migrationsToApply, config, createConnector)

	notifier := notifications.CreateNotifier(config)
	text := fmt.Sprintf("Migrations applied: %d", len(migrationsToApply))
	resp, err := notifier.Notify(text)

	if err != nil {
		log.Printf("Notifier err: %q", err)
	} else {
		log.Printf("Notifier response: %q", resp)
	}

	return migrationsToApply
}

// ExecuteMigrator is a function which executes actions on resources defined in config passed as first argument action defined as second argument
// and using connector created by a function passed as third argument and disk loader created by a function passed as fourth argument
func ExecuteMigrator(config *config.Config, action *string, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) int {

	switch *action {
	case PrintConfigAction:
		log.Printf("Configuration file ==>\n%v\n", config)
		return 0
	case ListDiskMigrationsAction:
		diskMigrations := LoadDiskMigrations(config, createLoader)
		log.Printf("List of disk migrations ==>\n%v", types.MigrationArrayString(diskMigrations))
		return 0
	case ListDBMigrationsAction:
		dbMigrations := LoadDBMigrations(config, createConnector)
		log.Printf("List of db migrations ==> \n%v", types.MigrationDBArrayString(dbMigrations))
		return 0
	case ListDBTenantsAction:
		dbTenants := LoadDBTenants(config, createConnector)
		log.Printf("List of db tenants ==> \n%v", types.TenantArrayString(dbTenants))
		return 0
	case ApplyAction:
		migrationsApplied := ApplyMigrations(config, createConnector, createLoader)
		log.Printf("List of migrations applied ==>\n%v", types.MigrationArrayString(migrationsApplied))
		return 0
	default:
		log.Printf("Unknown action to run %q. For usage please run migrator with -h flag.", *action)
		return 1
	}
}
