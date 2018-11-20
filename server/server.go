package server

import (
	"encoding/json"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/xcli"
	"log"
	"net/http"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func makeHandler(handler func(w http.ResponseWriter, r *http.Request, c *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader), config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config, createConnector, createLoader)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func diskMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	diskMigrations := xcli.LoadDiskMigrations(config, createLoader)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diskMigrations)
}

func dbTenantsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	dbTenants := xcli.LoadDBTenants(config, createConnector)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbTenants)
}

func dbMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	dbMigrations := xcli.LoadDBMigrations(config, createConnector)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbMigrations)
}

func applyHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	if r.Method != "POST" {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	migrationsApplied := xcli.ApplyMigrations(config, createConnector, createLoader)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(migrationsApplied)
}

func registerHandlers(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/config", makeHandler(configHandler, config, nil, nil))
	http.HandleFunc("/diskMigrations", makeHandler(diskMigrationsHandler, config, nil, createLoader))
	http.HandleFunc("/dbTenants", makeHandler(dbTenantsHandler, config, createConnector, nil))
	http.HandleFunc("/dbMigrations", makeHandler(dbMigrationsHandler, config, createConnector, nil))
	http.HandleFunc("/apply", makeHandler(applyHandler, config, createConnector, createLoader))
}

// Start starts simple Migrator API endpoint using config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func Start(config *config.Config) {
	registerHandlers(config, db.CreateConnector, loader.CreateLoader)
	log.Printf("Migrator web server starting on port %s...", config.Port)
	http.ListenAndServe(":"+config.Port, nil)
}
