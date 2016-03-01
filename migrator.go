package main

import (
	"flag"
	"fmt"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/disk"
	"github.com/lukaszbudnik/migrator/xcli"
	"os"
)

const (
	defaultConfigFile = "migrator.yaml"
)

func main() {
	validActions := []string{xcli.ApplyAction, xcli.ListDiskMigrationsAction, xcli.ListDBTenantsAction, xcli.ListDBMigrationsAction}

	configFile := flag.String("configFile", defaultConfigFile, "path to migrator.yaml")
	action := flag.String("action", xcli.ApplyAction, fmt.Sprintf("migrator action to apply, valid actions are: %q", validActions))
	verbose := flag.Bool("verbose", false, "set to \"true\" to print more data to output")
	flag.Parse()

	ret := xcli.ExecuteMigrator(configFile, action, verbose, db.CreateConnector, disk.CreateLoader)
	os.Exit(ret)
}
