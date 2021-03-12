package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/data"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

var (
	configFile          = "../test/migrator-postgresql.yaml"
	configFileOverrides = "../test/migrator-overrides.yaml"
)

func newTestRequestV1(method, url string, body io.Reader) (*http.Request, error) {
	versionURL := "/v1" + url
	return http.NewRequest(method, versionURL, body)
}

func newTestRequestV2(method, url string, body io.Reader) (*http.Request, error) {
	versionURL := "/v2" + url
	return http.NewRequest(method, versionURL, body)
}

func testSetupRouter(config *config.Config, newCoordinator func(ctx context.Context, config *config.Config) coordinator.Coordinator) *gin.Engine {
	versionInfo := &types.VersionInfo{Release: "GitBranch", CommitSha: "GitCommitSha", CommitDate: "2020-01-08T09:56:41+01:00", APIVersions: []types.APIVersion{types.APIV2}}
	gin.SetMode(gin.ReleaseMode)
	return SetupRouter(versionInfo, config, newCoordinator)
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

// section /

func TestRoot(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"release":"GitBranch","commitSha":"GitCommitSha","commitDate":"2020-01-08T09:56:41+01:00","apiVersions":["v2"]}`, strings.TrimSpace(w.Body.String()))
}

func TestRootWithPathPrefix(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	config.PathPrefix = "/migrator"

	router := testSetupRouter(config, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/migrator/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"release":"GitBranch","commitSha":"GitCommitSha","commitDate":"2020-01-08T09:56:41+01:00","apiVersions":["v2"]}`, strings.TrimSpace(w.Body.String()))
}

// /v1 API

func TestConfigRoute(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, nil)

	w := httptest.NewRecorder()
	req, _ := newTestRequestV1("GET", "/config", nil)
	router.ServeHTTP(w, req)

	// v1 is now removed
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// /v2 API

func TestGraphQLSchema(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, newMockedCoordinator)

	w := httptest.NewRecorder()
	req, _ := newTestRequestV2("GET", "/schema", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, strings.TrimSpace(data.SchemaDefinition), strings.TrimSpace(w.Body.String()))
}

func TestGraphQLQueryWithVariables(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, newMockedCoordinator)

	w := httptest.NewRecorder()
	req, _ := newTestRequestV2("POST", "/service", strings.NewReader(`
    {
      "query": "query SourceMigration($file: String!) { sourceMigration(file: $file) { name, migrationType, sourceDir, file } }",
      "operationName": "SourceMigration",
      "variables": { "file": "source/201602220001.sql" }
    }
  `))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"data":{"sourceMigration":{"name":"201602220001.sql","migrationType":"SingleMigration","sourceDir":"source","file":"source/201602220001.sql"}}}`, strings.TrimSpace(w.Body.String()))
}

func TestGraphQLQueryError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, newMockedCoordinator)

	w := httptest.NewRecorder()
	req, _ := newTestRequestV2("POST", "/service", strings.NewReader(`
    {
      "query": "query SourceMigration($file: String!) { sourceMigration(file: $file) { name, migrationType, sourceDir, file } }",
      "operationName": "SourceMigration",
    }
  `))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `{"error":"Invalid request, please see documentation for valid JSON payload"}`, strings.TrimSpace(w.Body.String()))
}
