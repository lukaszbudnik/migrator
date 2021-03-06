package main

import (
	"bytes"
	"context"
	"flag"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/server"
	"github.com/lukaszbudnik/migrator/types"
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

	common.Log("INFO", "migrator version %v, build %v, date %v", GitBranch, GitCommitSha, GitCommitDate)

	flag := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	buf := new(bytes.Buffer)
	flag.SetOutput(buf)

	var configFile string
	flag.StringVar(&configFile, "configFile", DefaultConfigFile, "path to migrator configuration yaml file")

	if err := flag.Parse(os.Args[1:]); err != nil {
		common.Log("ERROR", buf.String())
		os.Exit(1)
	}

	cfg, err := config.FromFile(configFile)
	if err != nil {
		common.Log("ERROR", "Error reading config file: %v", err)
		os.Exit(1)
	}

	var createCoordinator = func(ctx context.Context, config *config.Config) coordinator.Coordinator {
		coordinator := coordinator.New(ctx, config, db.New, loader.New, notifications.New)
		return coordinator
	}

	gin.SetMode(gin.ReleaseMode)
	versionInfo := &types.VersionInfo{Release: GitBranch, CommitSha: GitCommitSha, CommitDate: GitCommitDate, APIVersions: []string{"v1", "v2"}}
	g := server.SetupRouter(versionInfo, cfg, createCoordinator)
	if err := g.Run(":" + server.GetPort(cfg)); err != nil {
		common.Log("ERROR", "Error starting migrator: %v", err)
	}

}
