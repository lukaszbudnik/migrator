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

type requestIDKey struct{}

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
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
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
	requestID := r.Context().Value(requestIDKey{})
	if r.Method != http.MethodGet {
		log.Printf("[%v] ERROR /config - wrong method: %v", requestID, r.Method)
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%v] INFO /config", requestID)
	w.Header().Set("Content-Type", "application/x-yaml")
	fmt.Fprintf(w, "%v", config)
}

func diskMigrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	requestID := r.Context().Value(requestIDKey{})
	if r.Method != http.MethodGet {
		log.Printf("[%v] ERROR /diskMigrations - wrong method: %v", requestID, r.Method)
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%v] INFO /diskMigrations", requestID)
	loader := createAndInitLoader(config, newLoader)
	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		log.Printf("[%v] ERROR /diskMigrations - internal error: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	log.Printf("[%v] INFO /diskMigrations - returning disk migrations: %v", requestID, len(diskMigrations))
	jsonResponse(w, diskMigrations)
}

func migrationsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	requestID := r.Context().Value(requestIDKey{})
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		log.Printf("[%v] ERROR /migrations - wrong method: %v", requestID, r.Method)
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%v] INFO /migrations", requestID)
	if r.Method == http.MethodGet {
		migrationsGetHandler(w, r, config, newConnector, newLoader)
	}
	if r.Method == http.MethodPost {
		migrationsPostHandler(w, r, config, newConnector, newLoader)
	}
}

func migrationsGetHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	requestID := r.Context().Value(requestIDKey{})
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error creating connector: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	defer connector.Dispose()
	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error getting DB migrations: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	log.Printf("[%v] INFO /migrations - returning DB migrations: %v", requestID, len(dbMigrations))
	jsonResponse(w, dbMigrations)
}

func migrationsPostHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	requestID := r.Context().Value(requestIDKey{})
	loader := createAndInitLoader(config, newLoader)
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error creating connector: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	defer connector.Dispose()

	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error getting disk migrations: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error getting DB migrations: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)

	if !verified {
		log.Printf("[%v] ERROR /migrations - checksum verification failed for migrations: %v", requestID, len(offendingMigrations))
		errorResponse(w, http.StatusFailedDependency, struct {
			ErrorMessage        string
			OffendingMigrations []types.Migration
		}{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	migrationsToApply := migrations.ComputeMigrationsToApply(diskMigrations, dbMigrations)
	log.Printf("[%v] INFO /migrations - found migrations to apply: %d", requestID, len(migrationsToApply))

	err = connector.ApplyMigrations(migrationsToApply)
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error applying migrations: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}
	log.Printf("[%v] INFO /migrations - returning applied migrations: %v", requestID, len(migrationsToApply))
	jsonResponse(w, migrationsToApply)
}

func tenantsHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	requestID := r.Context().Value(requestIDKey{})
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		log.Printf("[%v] ERROR /tenants - wrong method: %v", requestID, r.Method)
		errorDefaultResponse(w, http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%v] INFO /tenants", requestID)

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
		errorInternalServerErrorResponse(w, err)
		return
	}
	defer connector.Dispose()
	tenants, err := connector.GetTenants()
	if err != nil {
		errorInternalServerErrorResponse(w, err)
		return
	}
	jsonResponse(w, tenants)
}

func tenantsPostHandler(w http.ResponseWriter, r *http.Request, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	requestID := r.Context().Value(requestIDKey{})

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
	var tenant struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(body, &tenant)
	if err != nil || tenant.Name == "" {
		errorResponseStatusErrorMessage(w, http.StatusBadRequest, "400 bad request")
		return
	}

	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error getting disk migrations: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error getting DB migrations: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)

	if !verified {
		log.Printf("[%v] ERROR /migrations - checksum verification failed for migrations: %v", requestID, len(offendingMigrations))
		errorResponse(w, http.StatusFailedDependency, struct {
			ErrorMessage        string
			OffendingMigrations []types.Migration
		}{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	// filter only tenant schemas
	migrationsToApply := migrations.FilterTenantMigrations(diskMigrations)
	log.Printf("Found migrations to apply: %d", len(migrationsToApply))

	err = connector.AddTenantAndApplyMigrations(tenant.Name, migrationsToApply)
	if err != nil {
		log.Printf("[%v] ERROR /migrations - internal error adding new tenant: %v", requestID, err.Error())
		errorInternalServerErrorResponse(w, err)
		return
	}

	text := fmt.Sprintf("Tenant %q added, migrations applied: %d", tenant.Name, len(migrationsToApply))
	sendNotification(config, text)

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
	log.Printf("INFO Migrator web server starting on port %s", port)

	router := registerHandlers(config, db.NewConnector, loader.NewLoader)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: tracing()(router),
	}

	err := server.ListenAndServe()

	return server, err
}
