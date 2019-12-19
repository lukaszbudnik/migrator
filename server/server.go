package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/migrations"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
)

const (
	defaultPort     string = "8080"
	requestIDHeader string = "X-Request-Id"
)

func getPort(config *config.Config) string {
	if len(strings.TrimSpace(config.Port)) == 0 {
		return defaultPort
	}
	return config.Port
}

func sendNotification(ctx context.Context, config *config.Config, text string) {
	notifier := notifications.NewNotifier(config)
	resp, err := notifier.Notify(text)

	if err != nil {
		common.LogError(ctx, "Notifier err: %v", err)
	} else {
		common.LogInfo(ctx, "Notifier response: %v", resp)
	}
}

func createAndInitLoader(config *config.Config, newLoader func(*config.Config) loader.Loader) loader.Loader {
	loader := newLoader(config)
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

func tracing() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// requestID
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			ctx := context.WithValue(r.Context(), common.RequestIDKey{}, requestID)
			// action
			action := fmt.Sprintf("%v %v", r.Method, r.RequestURI)
			ctx = context.WithValue(ctx, common.ActionKey{}, action)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func makeHandler(handler func(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader), config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config, newConnector, newLoader)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	if r.Method != http.MethodGet {
		common.LogError(r.Context(), "Wrong method")
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	common.LogInfo(r.Context(), "returning config file")
	w.Header().Set("Content-Type", "application/x-yaml")
	fmt.Fprintf(w, "%v", config)
}

func diskMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	if r.Method != http.MethodGet {
		common.LogError(r.Context(), "Wrong method")
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	common.LogInfo(r.Context(), "Start")
	loader := createAndInitLoader(config, newLoader)
	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		common.LogError(r.Context(), "Error getting disk migrations: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	common.LogInfo(r.Context(), "Returning disk migrations: %v", len(diskMigrations))
	jsonResponse(w, diskMigrations)
}

func migrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		common.LogError(r.Context(), "Wrong method: %v", r.Method)
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	common.LogInfo(r.Context(), "Start")
	if r.Method == http.MethodGet {
		migrationsGetHandler(w, r, config, newConnector, newLoader)
	}
	if r.Method == http.MethodPost {
		migrationsPostHandler(w, r, config, newConnector, newLoader)
	}
}

func migrationsGetHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(r.Context(), "Error creating connector: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	defer connector.Dispose()
	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		common.LogError(r.Context(), "Error getting DB migrations: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	common.LogInfo(r.Context(), "Returning DB migrations: %v", len(dbMigrations))
	jsonResponse(w, dbMigrations)
}

func migrationsPostHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	loader := createAndInitLoader(config, newLoader)
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(r.Context(), "Error creating connector: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	defer connector.Dispose()

	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		common.LogError(r.Context(), "Error getting disk migrations: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		common.LogError(r.Context(), "Error getting DB migrations: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)

	if !verified {
		common.LogError(r.Context(), "Checksum verification failed for migrations: %v", len(offendingMigrations))
		errorResponse(w, http.StatusFailedDependency, struct {
			ErrorMessage        string
			OffendingMigrations []types.Migration
		}{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	migrationsToApply := migrations.ComputeMigrationsToApply(r.Context(), diskMigrations, dbMigrations)
	common.LogInfo(r.Context(), "Found migrations to apply: %d", len(migrationsToApply))

	err = connector.ApplyMigrations(r.Context(), migrationsToApply)
	if err != nil {
		common.LogError(r.Context(), "Error applying migrations: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	text := fmt.Sprintf("Applied migrations: %v", len(migrationsToApply))
	sendNotification(r.Context(), config, text)

	common.LogInfo(r.Context(), "Returning applied migrations: %v", len(migrationsToApply))
	jsonResponse(w, migrationsToApply)
}

func tenantsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		common.LogError(r.Context(), "Wrong method")
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	common.LogInfo(r.Context(), "Start")

	if r.Method == http.MethodGet {
		tenantsGetHandler(w, r, config, newConnector, newLoader)
	}
	if r.Method == http.MethodPost {
		tenantsPostHandler(w, r, config, newConnector, newLoader)
	}
}

func tenantsGetHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(r.Context(), "Error creating connector: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	defer connector.Dispose()
	tenants, err := connector.GetTenants()
	if err != nil {
		common.LogError(r.Context(), "Error getting tenants: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	common.LogInfo(r.Context(), "Returning tenants: %v", len(tenants))
	jsonResponse(w, tenants)
}

func tenantsPostHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	loader := createAndInitLoader(config, newLoader)
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(r.Context(), "Error creating connector: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	defer connector.Dispose()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		common.LogError(r.Context(), "Error reading request: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	var tenant struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(body, &tenant)
	if err != nil || tenant.Name == "" {
		common.LogError(r.Context(), "Bad request: %v", err.Error())
		errorResponseStatusErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		common.LogError(r.Context(), "Error getting disk migrations: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		common.LogError(r.Context(), "Error getting DB migrations: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)

	if !verified {
		common.LogError(r.Context(), "Checksum verification failed for migrations: %v", len(offendingMigrations))
		errorResponse(w, http.StatusFailedDependency, struct {
			ErrorMessage        string
			OffendingMigrations []types.Migration
		}{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	// filter only tenant schemas
	migrationsToApply := migrations.FilterTenantMigrations(r.Context(), diskMigrations)
	common.LogInfo(r.Context(), "Found migrations to apply: %d", len(migrationsToApply))

	err = connector.AddTenantAndApplyMigrations(r.Context(), tenant.Name, migrationsToApply)
	if err != nil {
		common.LogError(r.Context(), "Error adding new tenant: %v", err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	text := fmt.Sprintf("Tenant %q added, migrations applied: %v", tenant.Name, len(migrationsToApply))
	sendNotification(r.Context(), config, text)

	common.LogInfo(r.Context(), text)
	jsonResponse(w, migrationsToApply)
}

func registerHandlers(config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", http.NotFoundHandler())
	router.Handle("/config", makeHandler(configHandler, config, nil, nil))
	router.Handle("/diskMigrations", makeHandler(diskMigrationsHandler, config, nil, newLoader))
	router.Handle("/migrations", makeHandler(migrationsHandler, config, newConnector, newLoader))
	router.Handle("/tenants", makeHandler(tenantsHandler, config, newConnector, newLoader))

	return router
}

// Start starts simple Migrator API endpoint using config passed as first argument
// and using connector created by a function passed as second argument and disk loader created by a function passed as third argument
func Start(config *config.Config) (*http.Server, error) {
	port := getPort(config)
	log.Printf("INFO migrator starting on http://0.0.0.0:%s", port)

	router := registerHandlers(config, db.NewConnector, loader.NewLoader)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: tracing()(router),
	}

	err := server.ListenAndServe()

	return server, err
}
