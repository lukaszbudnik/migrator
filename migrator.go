package main

import (
	"flag"
	"fmt"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/server"
	"github.com/lukaszbudnik/migrator/xcli"
	"os"
)

const (
	defaultConfigFile = "migrator.yaml"
	defaultMode       = "tool"
)

func main() {
	validActions := []string{xcli.ApplyAction, xcli.PrintConfigAction, xcli.ListDiskMigrationsAction, xcli.ListDBTenantsAction, xcli.ListDBMigrationsAction}

	configFile := flag.String("configFile", defaultConfigFile, "path to migrator.yaml")
	action := flag.String("action", xcli.ApplyAction, fmt.Sprintf("migrator action to apply, valid actions are: %q", validActions))
	mode := flag.String("mode", defaultMode, fmt.Sprintf("migrator mode to run: \"tool\" or \"server\""))
	flag.Parse()

	config := xcli.ReadConfig(configFile)

	if *mode == "tool" {
		ret := xcli.ExecuteMigrator(config, action, db.CreateConnector, loader.CreateLoader)
		os.Exit(ret)
	} else {
		server.Start(config, db.CreateConnector, loader.CreateLoader)
	}

}
