package xcli

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/disk"
	"github.com/lukaszbudnik/migrator/migrations"
	"github.com/lukaszbudnik/migrator/types"
	"log"
)

const (
	ApplyAction              = "apply"
	ListDBMigrationsAction   = "listDBMigrations"
	ListDBTenantsAction      = "listDBTenants"
	ListDiskMigrationsAction = "listDiskMigrations"
)

func ExecuteMigrator(configFile *string, action *string, verbose *bool, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) int {

	readConfig := func() *config.Config {
		config := config.FromFile(*configFile)

		log.Println("Read configuration file ==> OK")
		if *verbose {
			log.Printf("Configuration file ==>\n%v\n", config)
		}
		return config
	}

	loadDiskMigrations := func(loader disk.Loader) []types.Migration {
		diskMigrations := loader.GetDiskMigrations()
		log.Printf("Read disk migrations ==> OK")
		if *verbose || *action == ListDiskMigrationsAction {
			log.Printf("List of disk migrations ==>\n%v", types.MigrationArrayString(diskMigrations))
		}
		return diskMigrations
	}

	loadDBMigrations := func(connector db.Connector) []types.MigrationDB {
		dbMigrations := connector.GetMigrations()
		if *verbose || *action == ListDBMigrationsAction {
			log.Printf("List of db migrations ==> \n%v", types.MigrationDBArrayString(dbMigrations))
		}
		return dbMigrations
	}

	loadDBTenants := func(connector db.Connector) []string {
		dbTenants := connector.GetTenants()
		if *verbose || *action == ListDBTenantsAction {
			log.Printf("List of db tenants ==> \n%v", types.TenantArrayString(dbTenants))
		}
		return dbTenants
	}

	switch *action {
	case ListDiskMigrationsAction:
		config := readConfig()
		loader := createLoader(config)
		loadDiskMigrations(loader)
		return 0
	case ListDBMigrationsAction:
		config := readConfig()
		connector := createConnector(config)
		connector.Init()
		defer connector.Dispose()
		loadDBMigrations(connector)
		return 0
	case ListDBTenantsAction:
		config := readConfig()
		connector := createConnector(config)
		connector.Init()
		defer connector.Dispose()
		loadDBTenants(connector)
		return 0
	case ApplyAction:
		config := readConfig()
		loader := createLoader(config)
		diskMigrations := loadDiskMigrations(loader)
		connector := createConnector(config)
		connector.Init()
		defer connector.Dispose()
		dbMigrations := loadDBMigrations(connector)
		migrationsToApply := migrations.ComputeMigrationsToApply(diskMigrations, dbMigrations)
		if *verbose {
			log.Printf("List of migrations to apply ==>\n%v", types.MigrationArrayString(migrationsToApply))
		}
		connector.ApplyMigrations(migrationsToApply)
		return 0
	default:
		log.Printf("Unknown action to run %q. For usage please run migrator with -h flag.", *action)
		return 1
	}
}
