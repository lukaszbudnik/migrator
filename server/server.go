package server

// 90.8%

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/migrations"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
)

const (
	defaultPort string = "8080"
)

func getPort(config *config.Config) string {
	if len(strings.TrimSpace(config.Port)) == 0 {
		return defaultPort
	}
	return config.Port
}

func sendNotification(config *config.Config, text string) {
	notifier := notifications.NewNotifier(config)
	resp, err := notifier.Notify(text)

	if err != nil {
		log.Printf("Notifier err: %v", err)
	} else {
		log.Printf("Notifier response: %v", resp)
	}
}

func createAndInitLoader(config *config.Config, newLoader func(*config.Config) loader.Loader) loader.Loader {
	loader := newLoader(config)
	log.Println("Loader instance created")
	return loader
}

func createAndInitConnector(config *config.Config, newConnector func(*config.Config) (db.Connector, error)) (db.Connector, error) {
	connector, err := newConnector(config)
	if err != nil {
		return nil, err
	}
	if err := connector.Init(); err != nil {
		return nil, err
	}
	log.Println("Connector instance created")
	return connector, nil
}

func errorResponse(w http.ResponseWriter, errorStatus int, response interface{}) {
	w.WriteHeader(errorStatus)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func errorResponseStatusErrorMessage(w http.ResponseWriter, errorStatus int, errorMessage string) {
	errorResponse(w, errorStatus, struct{ ErrorMessage string }{errorMessage})
}

func errorDefaultResponse(w http.ResponseWriter, errorStatus int) {
	errorResponseStatusErrorMessage(w, errorStatus, http.StatusText(errorStatus))
}

func errorInternalServerErrorResponse(w http.ResponseWriter, err error) {
	errorResponseStatusErrorMessage(w, http.StatusInternalServerError, err.Error())
}

func jsonResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func makeHandler(handler func(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader), config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config, newConnector, newLoader)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	if r.Method != http.MethodGet {
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	fmt.Fprintf(w, "%v", config)
}

func diskMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	if r.Method != http.MethodGet {
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	loader := createAndInitLoader(config, newLoader)
	diskMigrations, err := getDiskMigrations(loader)
	if err != nil {
		errorInternalServerErrorResponse(w, err)
		return
	}
	jsonResponse(w, diskMigrations)
}

func migrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	switch r.Method {
	case http.MethodGet:
		connector, err := createAndInitConnector(config, newConnector)
		if err != nil {
			errorInternalServerErrorResponse(w, err)
			return
		}
		defer connector.Dispose()
		dbMigrations, err := getDBMigrations(connector)
		if err != nil {
			log.Printf("Error getting DB migrations: %v", err)
			errorInternalServerErrorResponse(w, err)
			return
		}
		log.Printf("Got DB migrations: %v", len(dbMigrations))
		jsonResponse(w, dbMigrations)
	case http.MethodPost:
		loader := createAndInitLoader(config, newLoader)
		connector, err := createAndInitConnector(config, newConnector)
		if err != nil {
			errorInternalServerErrorResponse(w, err)
			return
		}
		defer connector.Dispose()
		verified := verifyMigrations(w, config, connector, loader)
		if verified {
			migrationsApplied, err := applyMigrations(config, connector, loader)
			if err != nil {
				errorInternalServerErrorResponse(w, err)
				return
			}
			jsonResponse(w, migrationsApplied)
		}
	default:
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
	}
}

func tenantsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	switch r.Method {
	case http.MethodGet:
		connector, err := createAndInitConnector(config, newConnector)
		if err != nil {
			errorInternalServerErrorResponse(w, err)
			return
		}
		defer connector.Dispose()
		tenants, err := getDBTenants(connector)
		if err != nil {
			errorInternalServerErrorResponse(w, err)
			return
		}
		jsonResponse(w, tenants)
	case http.MethodPost:
		loader := createAndInitLoader(config, newLoader)
		connector, err := createAndInitConnector(config, newConnector)
		if err != nil {
			errorInternalServerErrorResponse(w, err)
			return
		}
		defer connector.Dispose()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errorInternalServerErrorResponse(w, err)
			return
		}
		var param struct {
			Name string `json:"name"`
		}
		err = json.Unmarshal(body, &param)
		if err != nil || param.Name == "" {
			errorResponseStatusErrorMessage(w, http.StatusBadRequest, "400 bad request")
			return
		}
		verified := verifyMigrations(w, config, connector, loader)
		if verified {
			migrationsApplied, err := addTenant(param.Name, config, connector, loader)
			if err != nil {
				errorResponseStatusErrorMessage(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonResponse(w, migrationsApplied)
		}
	default:
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
	}
}

func registerHandlers(config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	http.HandleFunc("/", makeHandler(configHandler, config, nil, nil))
	http.HandleFunc("/diskMigrations", makeHandler(diskMigrationsHandler, config, nil, newLoader))
	http.HandleFunc("/migrations", makeHandler(migrationsHandler, config, newConnector, newLoader))
	http.HandleFunc("/tenants", makeHandler(tenantsHandler, config, newConnector, newLoader))
}

// getDiskMigrations is a function which loads all migrations from disk as defined in config passed as first argument
// and using loader created by a function passed as second argument
func getDiskMigrations(loader loader.Loader) ([]types.Migration, error) {
	return loader.GetDiskMigrations()
}

// getDBTenants is a function which loads all tenants for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func getDBTenants(connector db.Connector) ([]string, error) {
	return connector.GetTenants()
}

// getDBMigrations is a function which loads all DB migrations for multi-tenant schemas from DB as defined in config passed as first argument
// and using connector created by a function passed as second argument
func getDBMigrations(connector db.Connector) ([]types.MigrationDB, error) {
	return connector.GetDBMigrations()
}

// applyMigrations is a function which applies disk migrations to DB as defined in config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func applyMigrations(config *config.Config, connector db.Connector, loader loader.Loader) (migrationsToApply []types.Migration, err error) {
	diskMigrations, err := getDiskMigrations(loader)
	if err != nil {
		return
	}
	log.Printf("Read disk migrations: %d", len(diskMigrations))

	dbMigrations, err := getDBMigrations(connector)
	if err != nil {
		return
	}
	log.Printf("Read DB migrations: %d", len(dbMigrations))

	migrationsToApply = migrations.ComputeMigrationsToApply(diskMigrations, dbMigrations)
	log.Printf("Found migrations to apply: %d", len(migrationsToApply))

	err = connector.ApplyMigrations(migrationsToApply)
	if err != nil {
		return
	}

	text := fmt.Sprintf("Migrations applied: %d", len(migrationsToApply))
	sendNotification(config, text)

	return
}

// addTenant creates new tenant in DB and applies all tenant migrations
func addTenant(tenant string, config *config.Config, connector db.Connector, loader loader.Loader) (migrationsToApply []types.Migration, err error) {

	diskMigrations, err := getDiskMigrations(loader)
	if err != nil {
		return
	}
	log.Printf("Read disk migrations: %d", len(diskMigrations))

	// filter only tenant schemas
	migrationsToApply = migrations.FilterTenantMigrations(diskMigrations)
	log.Printf("Found migrations to apply: %d", len(migrationsToApply))

	err = doAddTenantAndApplyMigrations(tenant, migrationsToApply, connector)
	if err != nil {
		return
	}

	text := fmt.Sprintf("Tenant %q added, migrations applied: %d", tenant, len(migrationsToApply))
	sendNotification(config, text)

	return
}

func doAddTenantAndApplyMigrations(tenant string, migrationsToApply []types.Migration, connector db.Connector) error {
	return connector.AddTenantAndApplyMigrations(tenant, migrationsToApply)
}

func verifyMigrations(w http.ResponseWriter, config *config.Config, connector db.Connector, loader loader.Loader) bool {
	verified, offendingMigrations, err := doVerifyMigrations(config, connector, loader)
	if err != nil {
		errorInternalServerErrorResponse(w, err)
		return false
	}
	if !verified {
		log.Printf("Checksum verification failed.")
		errorResponse(w, http.StatusFailedDependency, struct {
			ErrorMessage        string
			OffendingMigrations []types.Migration
		}{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
	}
	return verified
}

// doVerifyMigrations loads disk and db migrations and verifies their checksums
// see migrations.VerifyCheckSums for more information
func doVerifyMigrations(config *config.Config, connector db.Connector, loader loader.Loader) (bool, []types.Migration, error) {
	diskMigrations, err := getDiskMigrations(loader)
	if err != nil {
		return false, []types.Migration{}, err
	}

	dbMigrations, err := getDBMigrations(connector)
	if err != nil {
		return false, []types.Migration{}, err
	}
	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)
	return verified, offendingMigrations, nil
}

// Start starts simple Migrator API endpoint using config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func Start(config *config.Config) {
	registerHandlers(config, db.NewConnector, loader.NewLoader)
	port := getPort(config)
	log.Printf("Migrator web server starting on port %s", port)
	http.ListenAndServe(":"+port, nil)
}
