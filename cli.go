package main

import (
	"log"
	"os"
)

func executeMigrator(configFile *string, action *string, verbose *bool) {

	if *action != applyAction && *action != listDBMigrationsAction && *action != listDiskMigrationsAction {
		log.Panicf("Unknown action to run %#v. For usage please run migrator with -h flag.", *action)
	}

	config, err := readConfigFromFile(*configFile)
	if err != nil {
		log.Panicf("Could not read config file %q ==> %q", *configFile, err)
	}

	log.Println("Read configuration file ==> OK")
	if *verbose {
		log.Printf("Configuration file ==>\n%v\n", config)
	}

	allMigrations, err := listAllMigrations(*config)
	if err != nil {
		log.Fatalf("Failed to process migrations ==> %q", err)
	}

	log.Printf("Read all migrations ==> OK")

	if *verbose || *action == listDiskMigrationsAction {
		log.Printf("List of all disk migrations ==>\n%v\n", migrationDefinitionsString(allMigrations))
	}

	connector := CreateConnector(config)

	err = connector.Init()
	if err != nil {
		log.Fatalf("Failed to init DB connector ==> %q", err)
	}
	defer connector.Dispose()

	dbMigrations, err := connector.ListAllDBMigrations()
	if err != nil {
		log.Fatalf("Failed to read migrations from db ==> %q", err)
	}

	log.Println("Read all db migrations ==> OK")

	if *verbose || *action == listDBMigrationsAction {
		log.Printf("List of all db migrations ==> \n%v\n", dbMigrationsString(dbMigrations))
	}

	if *action != applyAction {
		os.Exit(0)
	}

	migrationsToApply := computeMigrationsToApply(allMigrations, dbMigrations)

	if *verbose {
		log.Printf("List of migrations to apply ==>\n%v\n", migrationDefinitionsString(migrationsToApply))
	}

	migrations, err := loadMigrations(*config, migrationsToApply)

	err = connector.ApplyMigrations(migrations)
	if err != nil {
		log.Fatalf("Failed to apply migrations to db ==> %q", err)
	}

}
