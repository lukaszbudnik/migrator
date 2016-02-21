package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultConfigFile        = "migrator.yaml"
	applyAction              = "apply"
	listDBMigrationsAction   = "listDBMigrations"
	listDiskMigrationsAction = "listDiskMigrations"
)

func main() {

	configFile := flag.String("configFile", defaultConfigFile, "path to migrator.yaml")
	action := flag.String("action", applyAction, fmt.Sprintf("migrator action to apply, valid actions are: %q", []string{applyAction, listDBMigrationsAction, listDiskMigrationsAction}))
	verbose := flag.Bool("verbose", false, "set to true/1 to print more data to output")
	flag.Parse()

	if *action != applyAction && *action != listDBMigrationsAction && *action != listDiskMigrationsAction {
		log.Fatalf("Unknown action to run %#v", *action)
	}

	config, err := readConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("Could not read config file %q ==> %q", *configFile, err)
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

	connector, err := CreateConnector(config.Driver)
	if err != nil {
		log.Fatalf("Failed to create DB connector ==> %q", err)
	}

	dbMigrations, err := connector.ListAllDBMigrations(*config)
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

	err = connector.ApplyMigrations(*config, migrations)
	if err != nil {
		log.Fatalf("Failed to apply migrations to db ==> %q", err)
	}

}
