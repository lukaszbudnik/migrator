package main

import (
	"log"
)

func executeMigrator(configFile *string, action *string, verbose *bool, createConnector func(*Config) Connector) int {

	readConfig := func() *Config {
		config := readConfigFromFile(*configFile)

		log.Println("Read configuration file ==> OK")
		if *verbose {
			log.Printf("Configuration file ==>\n%v\n", config)
		}
		return config
	}

	loadDiskMigrations := func(config *Config) []MigrationDefinition {
		diskMigrations := listDiskMigrations(*config)
		log.Printf("Read disk migrations ==> OK")
		if *verbose || *action == listDiskMigrationsAction {
			log.Printf("List of disk migrations ==>\n%v", migrationDefinitionsString(diskMigrations))
		}
		return diskMigrations
	}

	loadDBMigrations := func(connector Connector) []DBMigration {
		dbMigrations, _ := connector.ListAllDBMigrations()
		if *verbose || *action == listDBMigrationsAction {
			log.Printf("List of db migrations ==> \n%v", dbMigrationsString(dbMigrations))
		}
		return dbMigrations
	}

	switch *action {
	case listDiskMigrationsAction:
		config := readConfig()
		loadDiskMigrations(config)
		return 0
	case listDBMigrationsAction:
		config := readConfig()
		connector := createConnector(config)
		loadDBMigrations(connector)
		return 0
	case applyAction:
		config := readConfig()
		diskMigrations := loadDiskMigrations(config)
		connector := createConnector(config)
		dbMigrations := loadDBMigrations(connector)
		migrationsToApply := computeMigrationsToApply(diskMigrations, dbMigrations)
		if *verbose {
			log.Printf("List of migrations to apply ==>\n%v", migrationDefinitionsString(migrationsToApply))
		}
		migrations, err := loadMigrations(*config, migrationsToApply)
		err = connector.ApplyMigrations(migrations)
		if err != nil {
			log.Printf("Failed to apply migrations to db ==> %q", err)
		}
		return 0
	default:
		log.Printf("Unknown action to run %q. For usage please run migrator with -h flag.", *action)
		return 1
	}
}
