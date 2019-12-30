package server

import (
	"bytes"
	"io"
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

func newTestRequest(method, url string, body io.Reader) (*http.Request, error) {
	versionURL := "/v1" + url
	return http.NewRequest(method, versionURL, body)
}

func TestGetDefaultPort(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	assert.Equal(t, "8080", GetPort(config))
}

func TestGetDefaultPortOverrides(t *testing.T) {
	config, err := config.FromFile(configFileOverrides)
	assert.Nil(t, err)
	assert.Equal(t, "8811", GetPort(config))
}

// section /config

func TestConfigRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, nil)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/config", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/x-yaml; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, config.String(), strings.TrimSpace(w.Body.String()))
}

// section /migrations/source

func TestDiskMigrationsRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, newMockedCoordinator)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/migrations/source", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"name":"201602220000.sql","sourceDir":"source","file":"source/201602220000.sql","migrationType":1,"contents":"select abc","checkSum":""},{"name":"201602220001.sql","sourceDir":"source","file":"source/201602220001.sql","migrationType":2,"contents":"select def","checkSum":""}]`, strings.TrimSpace(w.Body.String()))
}

// section /migrations/applied

func TestAppliedMigrationsRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, newMockedCoordinator)

	req, _ := newTestRequest(http.MethodGet, "/migrations/applied", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"name":"201602220000.sql","sourceDir":"source","file":"source/201602220000.sql","migrationType":1,"contents":"","checkSum":"","schema":"source","appliedAt":"2016-02-22T16:41:01.000000123Z"}]`, strings.TrimSpace(w.Body.String()))
}

// section /migrations

func TestMigrationsPostRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, newMockedCoordinator)

	json := []byte(`{"mode": "apply", "response": "full"}`)
	req, _ := newTestRequest(http.MethodPost, "/migrations", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Contains(t, strings.TrimSpace(w.Body.String()), `[{"name":"201602220000.sql","sourceDir":"source","file":"source/201602220000.sql","migrationType":1,"contents":"select abc","checkSum":""},{"name":"201602220001.sql","sourceDir":"source","file":"source/201602220001.sql","migrationType":2,"contents":"select def","checkSum":""}]`)
}

// section /tenants

func TestTenantsGetRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, newMockedCoordinator)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/tenants", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `["a","b","c"]`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsPostRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, newMockedCoordinator)

	json := []byte(`{"name": "new_tenant", "response": "full", "mode":"dry-run"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Contains(t, strings.TrimSpace(w.Body.String()), `[{"name":"201602220001.sql","sourceDir":"source","file":"source/201602220001.sql","migrationType":2,"contents":"select def","checkSum":""}]`)
}

func TestTenantsPostRouteRequestError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, newMockedCoordinator)

	json := []byte(`{"tenant": "new_tenant"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Invalid request, please see documentation for valid JSON payload"}`, strings.TrimSpace(w.Body.String()))
}

// func TestTenantsPostRouteVerifyCheckSumError(t *testing.T) {
// 	config, err := config.FromFile(configFile)
// 	assert.Nil(t, err)
//
// 	router := SetupRouter(config, newMockedConnector, newBrokenCheckSumMockedDiskLoader)
//
// 	json := []byte(`{"name": "new_tenant"}`)
// 	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))
//
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)
//
// 	assert.Equal(t, http.StatusFailedDependency, w.Code)
// 	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
// 	assert.Contains(t, strings.TrimSpace(w.Body.String()), `"error":"Checksum verification failed. Please review offending migrations."`)
// }
//

func TestRouteError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := SetupRouter(config, newMockedErrorCoordinator(0))

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/migrations/source", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Disk Loader: threshold 0 reached"}`, strings.TrimSpace(w.Body.String()))
}
