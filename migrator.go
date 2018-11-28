package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/core"
	"github.com/lukaszbudnik/migrator/server"
	"github.com/lukaszbudnik/migrator/utils"
	"log"
	"os"
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	validActions := []string{core.ApplyAction, core.AddTenantAction, core.PrintConfigAction, core.ListDiskMigrationsAction, core.ListDBTenantsAction, core.ListDBMigrationsAction}
	validModes := []string{core.ToolMode, core.ServerMode}

	flag := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	buf := new(bytes.Buffer)
	flag.SetOutput(buf)

	configFile := flag.String("configFile", core.DefaultConfigFile, "path to migrator configuration yaml file")
	mode := flag.String("mode", core.ToolMode, fmt.Sprintf("migrator mode to run: %q", validModes))
	// below flags apply only when run in tool mode
	action := flag.String("action", core.ApplyAction, fmt.Sprintf("when run in tool mode, action to execute, valid actions are: %q", validActions))
	tenant := flag.String("tenant", "", fmt.Sprintf("when run in tool mode and action set to %q, specifies new tenant name", core.AddTenantAction))
	err := flag.Parse(os.Args[1:])

	if err != nil {
		log.Fatal(buf)
		os.Exit(1)
	}

	if !utils.Contains(validModes, mode) {
		log.Printf("Invalid mode: %v", *mode)
		flag.Usage()
		log.Fatal(buf)
		os.Exit(1)
	}

	config, err := config.FromFile(*configFile)

	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if *mode == "server" {
		server.Start(config)
	} else {

		if !utils.Contains(validActions, action) {
			log.Printf("Invalid action: %v", *action)
			flag.Usage()
			log.Fatal(buf)
			os.Exit(1)
		}

		core.ExecuteMigrator(config, *action, *tenant)
	}

}
