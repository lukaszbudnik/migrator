// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * migrations_test.go
// DB & Disk operations are mocked using xcli_mocks.go

package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
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

func TestServerTenantsGet(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/tenants", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, createMockedConnector, nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `["a","b","c"]`, strings.TrimSpace(w.Body.String()))
}

func TestServerTenantsPost(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/tenants", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, createMockedConnector, createMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc","CheckSum":""},{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":1,"Contents":"select def","CheckSum":""}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerTenantsPostBadRequest(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	// empty JSON payload
	json := []byte("")
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	handler := makeHandler(tenantsHandler, config, createMockedConnector, createMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerDiskMigrationsGet(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/diskMigrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(diskMigrationsHandler, config, nil, createMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc","CheckSum":""},{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":1,"Contents":"select def","CheckSum":""}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerMigrationsGet(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/migrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(migrationsHandler, config, createMockedConnector, nil)
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
	handler := makeHandler(migrationsHandler, config, createMockedConnector, createMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":1,"Contents":"select def","CheckSum":""}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerMigrationsMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {
		req, _ := http.NewRequest(httpMethod, "http://example.com/migrations", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(migrationsHandler, config, createMockedConnector, createMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}

}

func TestServerTenantMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {
		req, _ := http.NewRequest(httpMethod, "http://example.com/tenants", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(tenantsHandler, config, createMockedConnector, createMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}

}

func TestServerDiskMigrationsMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {

		req, _ := http.NewRequest(httpMethod, "http://example.com/diskMigrations", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(diskMigrationsHandler, config, createMockedConnector, createMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}

}

func TestServerConfigMethodNotAllowed(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	httpMethods := []string{http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

	for _, httpMethod := range httpMethods {

		req, _ := http.NewRequest(httpMethod, "http://example.com/", nil)

		w := httptest.NewRecorder()
		handler := makeHandler(configHandler, config, createMockedConnector, createMockedDiskLoader)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	}

}
