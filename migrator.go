package main

import (
	"flag"
	"log"
)

const (
	defaultConfigFile = "migrator.yaml"
)

func main() {

	configFile := flag.String("configFile", defaultConfigFile, "path to migrator.yaml")
	verbose := flag.Bool("verbose", false, "prints more data to output")
	flag.Parse()

	config, err := readConfigFromFile(*configFile)
	if err != nil {
		log.Fatalf("Could not read config file %q ==> %q", *configFile, err)
	}

	log.Printf("Read configuration file ==> OK")
	if *verbose {
		log.Printf("Read configuration file ==> %#v", config)
	}

	allMigrations, err := listAllMigrations(*config)
	if err != nil {
		log.Fatalf("Failed to process migrations ==> %q", err)
	}

	log.Printf("Read all migrations ==> OK")

	if *verbose {
		log.Printf("Read all migrations ==> %#v", allMigrations)
	}

	dbMigrations, err := listAllDBMigrations(*config)
	if err != nil {
		log.Fatalf("Failed to read migrations from db ==> %q", err)
	}

	log.Printf("Read all db migrations ==> OK")

	if *verbose {
		log.Printf("Read all db migrations ==> %#v", dbMigrations)
	}

	migrationDefs := computeMigrationsToApply(allMigrations, dbMigrations)

	if *verbose {
		log.Printf("Migrations to apply ==> %#v", migrationDefs)
	}

	migrations, err := loadMigrations(*config, migrationDefs)

	err = applyMigrations(*config, migrations)
	if err != nil {
		log.Fatalf("Failed to apply migrations to db ==> %q", err)
	}

}
