package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/server"
	"github.com/lukaszbudnik/migrator/xcli"
	"log"
	"os"
)

func contains(slice []string, element *string) bool {
	for _, a := range slice {
		if a == *element {
			return true
		}
	}
	return false
}

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	validActions := []string{xcli.ApplyAction, xcli.PrintConfigAction, xcli.ListDiskMigrationsAction, xcli.ListDBTenantsAction, xcli.ListDBMigrationsAction}
	validModes := []string{xcli.ToolMode, xcli.ServerMode}

	flag := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	buf := new(bytes.Buffer)
	flag.SetOutput(buf)

	configFile := flag.String("configFile", xcli.DefaultConfigFile, "path to migrator configuration yaml file")
	action := flag.String("action", xcli.ApplyAction, fmt.Sprintf("migrator action to apply, valid actions are: %q", validActions))
	mode := flag.String("mode", xcli.ToolMode, fmt.Sprintf("migrator mode to run: %q", validModes))
	err := flag.Parse(os.Args[1:])

	if err != nil {
		log.Fatal(buf)
		os.Exit(1)
	}

	if !contains(validActions, action) {
		log.Printf("Invalid action: %v", *action)
		flag.Usage()
		log.Fatal(buf)
		os.Exit(1)
	}

	if !contains(validModes, mode) {
		log.Printf("Invalid mode: %v", *mode)
		flag.Usage()
		log.Fatal(buf)
		os.Exit(1)
	}

	config, err := config.FromFile(*configFile)

	if err == nil {
		if *mode == "server" {
			server.Start(config)
		} else {
			xcli.ExecuteMigrator(config, action)
		}
	}

}
