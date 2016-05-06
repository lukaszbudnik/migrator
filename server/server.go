package server

import (
	"encoding/json"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/disk"
	"github.com/lukaszbudnik/migrator/xcli"
	"log"
	"net/http"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func makeHandler(handler func(w http.ResponseWriter, r *http.Request, c *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader), config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config, createConnector, createLoader)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func diskMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) {
	diskMigrations := xcli.LoadDiskMigrations(config, createLoader)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diskMigrations)
}

func dbTenantsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) {
	dbTenants := xcli.LoadDBTenants(config, createConnector)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbTenants)
}

func dbMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) {
	dbMigrations := xcli.LoadDBMigrations(config, createConnector)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbMigrations)
}

func applyHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) {
	if r.Method != "POST" {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	migrationsApplied := xcli.ApplyMigrations(config, createConnector, createLoader)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(migrationsApplied)
}

func registerHandlers(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/config", makeHandler(configHandler, config, nil, nil))
	http.HandleFunc("/diskMigrations", makeHandler(diskMigrationsHandler, config, nil, createLoader))
	http.HandleFunc("/dbTenants", makeHandler(dbTenantsHandler, config, createConnector, nil))
	http.HandleFunc("/dbMigrations", makeHandler(dbMigrationsHandler, config, createConnector, nil))
	http.HandleFunc("/apply", makeHandler(applyHandler, config, createConnector, createLoader))
}

func Start(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) disk.Loader) {
	registerHandlers(config, createConnector, createLoader)
	log.Printf("Migrator web server starting...")
	http.ListenAndServe(":"+config.Port, nil)
}
