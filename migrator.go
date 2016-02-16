package main

import (
	"flag"
	"log"
)

const (
	defaultConfigFile = "migrator.yaml"
)

func main() {

	configFilePtr := flag.String("configFile", defaultConfigFile, "path to migrator.yaml")
	flag.Parse()

	config, err := readConfigFromFile(*configFilePtr)
	if err != nil {
		log.Fatalf("Could not read config file %q ==> %q", *configFilePtr, err)
	}

	allMigrations, err := listAllMigrations(*config)
	if err != nil {
		log.Fatalf("Failed to process migrations ==> %q", err)
	}

	log.Printf("Read all migrations %q", allMigrations)

	dbMigrations, err := listAllDBMigrations(*config)
	if err != nil {
		log.Fatalf("Failed to read migrations from db ==> %q", err)
	}

	log.Printf("Read all db migrations %v", dbMigrations)

}
