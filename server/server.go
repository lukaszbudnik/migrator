package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/core"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/types"
)

const (
	defaultPort string = "8080"
)

type tenantParam struct {
	Name string `json:"name"`
}

func getPort(config *config.Config) string {
	if len(strings.TrimSpace(config.Port)) == 0 {
		return defaultPort
	}
	return config.Port
}

func errorResponse(w http.ResponseWriter, errorStatus int, response interface{}) {
	w.WriteHeader(errorStatus)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func errorResponseStatusErrorMessage(w http.ResponseWriter, errorStatus int, errorMessage string) {
	errorResponse(w, errorStatus, struct{ ErrorMessage string }{errorMessage})
}

func errorMethodNotAllowedResponse(w http.ResponseWriter) {
	errorResponseStatusErrorMessage(w, http.StatusMethodNotAllowed, "405 method not allowed")
}

func jsonResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	if r.Method != http.MethodGet {
		errorMethodNotAllowedResponse(w)
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	fmt.Fprintf(w, "%v", config)
}

func diskMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {
	if r.Method != http.MethodGet {
		errorMethodNotAllowedResponse(w)
		return
	}
	diskMigrations := core.GetDiskMigrations(config, createLoader)
	jsonResponse(w, diskMigrations)
}

func migrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {

	switch r.Method {
	case http.MethodGet:
		dbMigrations := core.GetDBMigrations(config, createConnector)
		jsonResponse(w, dbMigrations)
	case http.MethodPost:
		verified, offendingMigrations := core.VerifyMigrations(config, createConnector, createLoader)
		if !verified {
			log.Printf("Checksum verification failed.")
			errorResponse(w, http.StatusFailedDependency, struct {
				ErrorMessage        string
				OffendingMigrations []types.Migration
			}{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		} else {
			migrationsApplied := core.ApplyMigrations(config, createConnector, createLoader)
			jsonResponse(w, migrationsApplied)
		}
	default:
		errorMethodNotAllowedResponse(w)
	}

}

func tenantsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, createConnector func(*config.Config) db.Connector, createLoader func(*config.Config) loader.Loader) {

	switch r.Method {
	case http.MethodGet:
		tenants := core.GetDBTenants(config, createConnector)
		jsonResponse(w, tenants)
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errorResponseStatusErrorMessage(w, http.StatusInternalServerError, "500 internal server error")
			return
		}
		var param tenantParam
		err = json.Unmarshal(body, &param)
		if err != nil || param.Name == "" {
			errorResponseStatusErrorMessage(w, http.StatusBadRequest, "400 bad request")
			return
		}
		verified, offendingMigrations := core.VerifyMigrations(config, createConnector, createLoader)
		if !verified {
			log.Printf("Checksum verification failed.")
			errorResponse(w, http.StatusFailedDependency, struct {
				ErrorMessage        string
				OffendingMigrations []types.Migration
			}{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		} else {
			migrationsApplied := core.AddTenant(param.Name, config, createConnector, createLoader)
			jsonResponse(w, migrationsApplied)
		}
	default:
		errorMethodNotAllowedResponse(w)
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
	port := getPort(config)
	log.Printf("Migrator web server starting on port %s", port)
	http.ListenAndServe(":"+port, nil)
}
