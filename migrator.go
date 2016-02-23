package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	defaultConfigFile        = "migrator.yaml"
	applyAction              = "apply"
	listDBMigrationsAction   = "listDBMigrations"
	listDBTenantsAction      = "listDBTenants"
	listDiskMigrationsAction = "listDiskMigrations"
)

func main() {
	validActions := []string{applyAction, listDiskMigrationsAction, listDBTenantsAction, listDBMigrationsAction}

	configFile := flag.String("configFile", defaultConfigFile, "path to migrator.yaml")
	action := flag.String("action", applyAction, fmt.Sprintf("migrator action to apply, valid actions are: %q", validActions))
	verbose := flag.Bool("verbose", false, "set to \"true\" to print more data to output")
	flag.Parse()

	ret := executeMigrator(configFile, action, verbose, CreateConnector, CreateLoader)
	os.Exit(ret)
}
