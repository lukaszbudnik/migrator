package server

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/core"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"log"
	"net/http"
	"strings"
)

const (
	defaultPort string = "8080"
)

type tenantParam struct {
	Name string `json:"name"`
}

func getDefaultPort(config *config.Config) string {
	if len(strings.TrimSpace(config.Port)) == 0 {
		return defaultPort
	}
	return config.Port
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func makeHandler(handler func(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader), config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config, createConnector, createLoader)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	if r.Method != "GET" {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	fmt.Fprintf(w, "%v", config)
}

func diskMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	if r.Method != "GET" {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	diskMigrations := core.LoadDiskMigrations(config, createLoader)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diskMigrations)
}

func migrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {

	switch r.Method {
	case http.MethodGet:
		dbMigrations := core.LoadDBMigrations(config, createConnector)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dbMigrations)
	case http.MethodPost:
		migrationsApplied := core.ApplyMigrations(config, createConnector, createLoader)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(migrationsApplied)
	default:
	    http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	}

}

func tenantsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {

	switch r.Method {
	case http.MethodGet:
			tenants := core.LoadDBTenants(config, createConnector)
	    w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tenants)
	case http.MethodPost:
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
					http.Error(w, "500 internal server error", http.StatusInternalServerError)
					return
			}
			var param tenantParam
			err = json.Unmarshal(body, &param)
			if err != nil || param.Name == "" {
					http.Error(w, "400 bad request", http.StatusBadRequest)
					return
			}
			migrationsApplied := core.AddTenant(param.Name, config, createConnector, createLoader)
	    w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(migrationsApplied)
	default:
	    http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	}

}

func registerHandlers(config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	http.HandleFunc("/", makeHandler(configHandler, config, nil, nil))
	http.HandleFunc("/diskMigrations", makeHandler(diskMigrationsHandler, config, nil, createLoader))
	http.HandleFunc("/migrations", makeHandler(migrationsHandler, config, createConnector, createLoader))
	http.HandleFunc("/tenants", makeHandler(tenantsHandler, config, createConnector, createLoader))
}

// Start starts simple Migrator API endpoint using config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func Start(config *config.Config) {
	registerHandlers(config, db.CreateConnector, loader.CreateLoader)
	port := getDefaultPort(config)
	log.Printf("Migrator web server starting on port %s", port)
	http.ListenAndServe(":"+port, nil)
}
