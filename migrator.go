package main

import (
	"bytes"
	"flag"
	"log"
	"os"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/server"
)

const (
	// DefaultConfigFile defines default file name of migrator configuration file
	DefaultConfigFile = "migrator.yaml"
)

// GitBranch stores git branch/tag, value injected during production build
var GitBranch string

// GitCommitSha stores git commit sha, value injected during production build
var GitCommitSha string

// GitCommitDate stores git commit date time, value injected during production build
var GitCommitDate string

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)

	log.Printf("INFO migrator version %v, build %v, date %v", GitBranch, GitCommitSha, GitCommitDate)

	flag := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	buf := new(bytes.Buffer)
	flag.SetOutput(buf)

	var configFile string
	flag.StringVar(&configFile, "configFile", DefaultConfigFile, "path to migrator configuration yaml file")
	err := flag.Parse(os.Args[1:])

	if err != nil {
		log.Fatal(buf)
		os.Exit(1)
	}

	config, err := config.FromFile(configFile)
	if err != nil {
		log.Fatalf("ERROR Error reading config file: %v", err)
	}

	srv, err := server.Start(config)
	if err != nil {
		log.Fatalf("ERROR Error starting: %v", err)
	}
	defer srv.Close()

}
