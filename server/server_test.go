package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/stretchr/testify/assert"
)

var (
	configFile          = "../test/migrator.yaml"
	configFileOverrides = "../test/migrator-overrides.yaml"
)

func TestGetDefaultPort(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	assert.Equal(t, "8080", getPort(config))
}

func TestGetDefaultPortOverrides(t *testing.T) {
	config, err := config.FromFile(configFileOverrides)
	assert.Nil(t, err)
	assert.Equal(t, "8811", getPort(config))
}

func TestRegisterHandlers(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	registerHandlers(config, nil, nil)
}

func TestServerDefaultHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/qwq", nil)

	w := httptest.NewRecorder()
	defaultHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// section: /config
func TestServerConfig(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(configHandler, config, nil, nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/x-yaml", w.HeaderMap["Content-Type"][0])
}

func TestServerConfigMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {

		req, _ := http.NewRequest(httpMethod, "http://example.com/", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(configHandler, config, newMockedConnector, newMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}
}

// section: /tenants
func TestServerTenantsGet(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/tenants", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, newMockedConnector, nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `["a","b","c"]`, strings.TrimSpace(w.Body.String()))
}

// func TestServerTenantsGetErrorConnector(t *testing.T) {
// 	config, err := config.FromFile(configFile)
// 	assert.Nil(t, err)
//
// 	req, _ := http.NewRequest(http.MethodGet, "http://example.com/tenants", nil)
//
// 	w := httptest.NewRecorder()
//
// 	handler := makeHandler(tenantsHandler, config, newConnectorReturnError, nil)
// 	handler(w, req)
//
// 	assert.Equal(t, http.StatusInternalServerError, w.Code)
// 	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
// 	assert.Equal(t, `{"ErrorMessage":"trouble maker"}`, strings.TrimSpace(w.Body.String()))
// }
//
// func TestServerTenantsGetErrorDBInit(t *testing.T) {
// 	config, err := config.FromFile(configFile)
// 	assert.Nil(t, err)
//
// 	req, _ := http.NewRequest(http.MethodGet, "http://example.com/tenants", nil)
//
// 	w := httptest.NewRecorder()
//
// 	handler := makeHandler(tenantsHandler, config, newMockedErrorConnector, nil)
// 	handler(w, req)
//
// 	assert.Equal(t, http.StatusInternalServerError, w.Code)
// 	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
// 	assert.Equal(t, `{"ErrorMessage":"trouble maker"}`, strings.TrimSpace(w.Body.String()))
// }

func TestServerTenantsPost(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/tenants", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, newMockedConnector, newMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":2,"Contents":"select def","CheckSum":""}]`, strings.TrimSpace(w.Body.String()))
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("trouble maker")
}

func TestServerTenantsPostIOError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPost, "http://example.com/tenants", errReader(0))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, newMockedConnector, newMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"ErrorMessage":"trouble maker"}`, strings.TrimSpace(w.Body.String()))
}

func TestServerTenantsPostBadRequest(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	// empty JSON payload
	json := []byte("")
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, newMockedConnector, newMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerTenantsPostFailedDependency(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, newMockedConnector, newBrokenCheckSumMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusFailedDependency, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"ErrorMessage":"Checksum verification failed. Please review offending migrations.","OffendingMigrations":[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc","CheckSum":"xxx"}]}`, strings.TrimSpace(w.Body.String()))
}

func TestServerTenantMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {
		req, _ := http.NewRequest(httpMethod, "http://example.com/tenants", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(tenantsHandler, config, newMockedConnector, newMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}

}

// section: /diskMigrations

func TestServerDiskMigrationsGet(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/diskMigrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(diskMigrationsHandler, config, nil, newMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc","CheckSum":""},{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":2,"Contents":"select def","CheckSum":""}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerDiskMigrationsMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {

		req, _ := http.NewRequest(httpMethod, "http://example.com/diskMigrations", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(diskMigrationsHandler, config, newMockedConnector, newMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}

}

// section: /migrations

func TestServerMigrationsGet(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/migrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(migrationsHandler, config, newMockedConnector, nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"","CheckSum":"","Schema":"source","Created":"2016-02-22T16:41:01.000000123Z"}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerMigrationsPost(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPost, "http://example.com/migrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(migrationsHandler, config, newMockedConnector, newMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":2,"Contents":"select def","CheckSum":""}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerMigrationsPostFailedDependency(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodPost, "http://example.com/migrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(migrationsHandler, config, newMockedConnector, newBrokenCheckSumMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusFailedDependency, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"ErrorMessage":"Checksum verification failed. Please review offending migrations.","OffendingMigrations":[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc","CheckSum":"xxx"}]}`, strings.TrimSpace(w.Body.String()))
}

func TestServerMigrationsMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {
		req, _ := http.NewRequest(httpMethod, "http://example.com/migrations", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(migrationsHandler, config, newMockedConnector, newMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}

}

func TestServerInternalServerErrors(t *testing.T) {
	c, err := config.FromFile(configFile)
	assert.Nil(t, err)

	requests := []struct {
		method          string
		path            string
		handler         func(http.ResponseWriter, *http.Request, *config.Config, func(*config.Config) (db.Connector, error), func(*config.Config) loader.Loader)
		createConnector func(config *config.Config) (db.Connector, error)
		createLoader    func(config *config.Config) loader.Loader
		payload         io.Reader
	}{{http.MethodGet, "tenants", tenantsHandler, newConnectorReturnError, newMockedDiskLoader, nil},
		{http.MethodGet, "tenants", tenantsHandler, newMockedErrorConnector(0), newMockedDiskLoader, nil},
		{http.MethodGet, "tenants", tenantsHandler, newMockedErrorConnector(1), newMockedDiskLoader, nil},
		{http.MethodPost, "tenants", tenantsHandler, newMockedErrorConnector(0), newMockedDiskLoader, bytes.NewBuffer([]byte(`{"name": "new_tenant"}`))},
		{http.MethodPost, "tenants", tenantsHandler, newMockedErrorConnector(1), newMockedDiskLoader, bytes.NewBuffer([]byte(`{"name": "new_tenant"}`))},
		{http.MethodPost, "tenants", tenantsHandler, newMockedErrorConnector(2), newMockedDiskLoader, bytes.NewBuffer([]byte(`{"name": "new_tenant"}`))},
		{http.MethodPost, "tenants", tenantsHandler, newMockedErrorConnector(2), newMockedErrorDiskLoader(1), bytes.NewBuffer([]byte(`{"name": "new_tenant"}`))},
		{http.MethodGet, "migrations", migrationsHandler, newConnectorReturnError, newMockedDiskLoader, nil},
		{http.MethodGet, "migrations", migrationsHandler, newMockedErrorConnector(0), newMockedDiskLoader, nil},
		{http.MethodGet, "migrations", migrationsHandler, newMockedErrorConnector(1), newMockedDiskLoader, nil},
		{http.MethodPost, "migrations", migrationsHandler, newMockedErrorConnector(0), newMockedDiskLoader, nil},
		{http.MethodPost, "migrations", migrationsHandler, newMockedErrorConnector(1), newMockedDiskLoader, nil},
		{http.MethodPost, "migrations", migrationsHandler, newMockedErrorConnector(2), newMockedDiskLoader, nil},
		{http.MethodPost, "migrations", migrationsHandler, newMockedErrorConnector(2), newMockedErrorDiskLoader(0), nil},
		{http.MethodPost, "migrations", migrationsHandler, newMockedErrorConnector(2), newMockedErrorDiskLoader(1), nil},
		{http.MethodPost, "migrations", migrationsHandler, newMockedErrorConnector(3), newMockedDiskLoader, nil},
		{http.MethodGet, "diskMigrations", diskMigrationsHandler, newMockedConnector, newMockedErrorDiskLoader(0), nil}}

	for _, r := range requests {
		req, _ := http.NewRequest(r.method, fmt.Sprintf("http://example.com/%v", r.path), r.payload)

		w := httptest.NewRecorder()
		handler := makeHandler(r.handler, c, r.createConnector, r.createLoader)
		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}
