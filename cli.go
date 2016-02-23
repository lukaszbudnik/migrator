package main

import (
	"log"
)

func executeMigrator(configFile *string, action *string, verbose *bool, createConnector func(*Config) Connector, createLoader func(*Config) Loader) int {

	readConfig := func() *Config {
		config := readConfigFromFile(*configFile)

		log.Println("Read configuration file ==> OK")
		if *verbose {
			log.Printf("Configuration file ==>\n%v\n", config)
		}
		return config
	}

	loadDiskMigrations := func(loader Loader) []Migration {
		diskMigrations := loader.GetDiskMigrations()
		log.Printf("Read disk migrations ==> OK")
		if *verbose || *action == listDiskMigrationsAction {
			log.Printf("List of disk migrations ==>\n%v", migrationsString(diskMigrations))
		}
		return diskMigrations
	}

	loadDBMigrations := func(connector Connector) []DBMigration {
		dbMigrations := connector.GetDBMigrations()
		if *verbose || *action == listDBMigrationsAction {
			log.Printf("List of db migrations ==> \n%v", dbMigrationsString(dbMigrations))
		}
		return dbMigrations
	}

	switch *action {
	case listDiskMigrationsAction:
		config := readConfig()
		loader := createLoader(config)
		loadDiskMigrations(loader)
		return 0
	case listDBMigrationsAction:
		config := readConfig()
		connector := createConnector(config)
		loadDBMigrations(connector)
		return 0
	case applyAction:
		config := readConfig()
		loader := createLoader(config)
		connector := createConnector(config)
		diskMigrations := loadDiskMigrations(loader)
		dbMigrations := loadDBMigrations(connector)
		migrationsToApply := computeMigrationsToApply(diskMigrations, dbMigrations)
		if *verbose {
			log.Printf("List of migrations to apply ==>\n%v", migrationsString(migrationsToApply))
		}
		err := connector.ApplyMigrations(migrationsToApply)
		if err != nil {
			log.Printf("Failed to apply migrations to db ==> %q", err)
		}
		return 0
	default:
		log.Printf("Unknown action to run %q. For usage please run migrator with -h flag.", *action)
		return 1
	}
}
