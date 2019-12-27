package server

import (
	"bytes"
	"errors"
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

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("trouble maker")
}

func newTestRequest(method, url string, body io.Reader) (*http.Request, error) {
	versionURL := "/v1" + url
	return http.NewRequest(method, versionURL, body)
}

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

// section /config

func TestConfigRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, nil, nil)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/config", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/x-yaml; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, config.String(), strings.TrimSpace(w.Body.String()))
}

// section /diskMigrations

func TestDiskMigrationsRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, nil, newMockedDiskLoader)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/diskMigrations", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc","CheckSum":""},{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":2,"Contents":"select def","CheckSum":""}]`, strings.TrimSpace(w.Body.String()))
}

func TestDiskMigrationsRouteError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, nil, newMockedErrorDiskLoader(0))

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/diskMigrations", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Disk Loader: threshold 0 reached"}`, strings.TrimSpace(w.Body.String()))
}

// section /tenants

func TestTenantsGetRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, nil)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/tenants", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `["a","b","c"]`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsGetRouteInitConnectorError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedErrorConnector(0), nil)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/tenants", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Connector: threshold 0 reached"}`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsGetRouteGetTenantsError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedErrorConnector(1), nil)

	w := httptest.NewRecorder()
	req, _ := newTestRequest("GET", "/tenants", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Connector: threshold 1 reached"}`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsPostRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newMockedDiskLoader)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Contains(t, strings.TrimSpace(w.Body.String()), `[{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":2,"Contents":"select def","CheckSum":""}]`)
}

func TestTenantsPostRouteRequestError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newMockedDiskLoader)

	req, _ := newTestRequest(http.MethodPost, "/tenants", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"invalid request"}`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsPostRouteInitConnectorError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedErrorConnector(0), newMockedDiskLoader)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Connector: threshold 0 reached"}`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsPostRouteGetDBMigrationsError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedErrorConnector(1), newMockedDiskLoader)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Connector: threshold 1 reached"}`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsPostRouteGetDiskMigrationsError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newMockedErrorDiskLoader(0))

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Disk Loader: threshold 0 reached"}`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsPostRouteAddTenantAndApplyMigrationsError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedErrorConnector(2), newMockedDiskLoader)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Mocked Error Connector: threshold 2 reached"}`, strings.TrimSpace(w.Body.String()))
}

func TestTenantsPostRouteVerifyCheckSumError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newBrokenCheckSumMockedDiskLoader)

	json := []byte(`{"name": "new_tenant"}`)
	req, _ := newTestRequest(http.MethodPost, "/tenants", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFailedDependency, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Contains(t, strings.TrimSpace(w.Body.String()), `"error":"Checksum verification failed. Please review offending migrations."`)
}

// section /migrations

func TestMigrationsGetRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newBrokenCheckSumMockedDiskLoader)

	req, _ := newTestRequest(http.MethodGet, "/migrations", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"","CheckSum":"","Schema":"source","Created":"2016-02-22T16:41:01.000000123Z"}]`, strings.TrimSpace(w.Body.String()))
}

func TestMigrationsPostRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newMockedDiskLoader)

	req, _ := newTestRequest(http.MethodPost, "/migrations", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Contains(t, strings.TrimSpace(w.Body.String()), `[{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":2,"Contents":"select def","CheckSum":""}]`)
}

func TestMigrationsPostRequestRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newMockedDiskLoader)

	json := []byte(`{"mode": "apply", "response": "summary"}`)
	req, _ := newTestRequest(http.MethodPost, "/migrations", bytes.NewBuffer(json))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Contains(t, strings.TrimSpace(w.Body.String()), `[{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":2,"Contents":"select def","CheckSum":""}]`)
}

func TestServerMigrationsPostFailedDependency(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := setupRouter(config, newMockedConnector, newBrokenCheckSumMockedDiskLoader)

	req, _ := newTestRequest(http.MethodPost, "/migrations", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFailedDependency, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Checksum verification failed. Please review offending migrations.","details":[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc","CheckSum":"xxx"}]}`, strings.TrimSpace(w.Body.String()))
}
