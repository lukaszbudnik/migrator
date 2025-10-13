package server

import (
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

func testSetupRouter(config *config.Config, newCoordinator coordinator.Factory) *gin.Engine {
	versionInfo := &types.VersionInfo{Release: "GitRef", Sha: "GitSha", APIVersions: []types.APIVersion{types.APIV2}}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	return SetupRouter(r, versionInfo, config, newNoopMetrics(), newCoordinator)
}

func testSetupRouterInDebug(config *config.Config, newCoordinator coordinator.Factory) *gin.Engine {
	versionInfo := &types.VersionInfo{Release: "GitRef", Sha: "GitSha", APIVersions: []types.APIVersion{types.APIV2}}
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	return SetupRouter(r, versionInfo, config, newNoopMetrics(), newCoordinator)
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
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Equal(t, `{"release":"GitRef","sha":"GitSha","apiVersions":["v2"]}`, strings.TrimSpace(w.Body.String()))
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
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Equal(t, `{"release":"GitRef","sha":"GitSha","apiVersions":["v2"]}`, strings.TrimSpace(w.Body.String()))
}

// /metrics
func TestCreateRouterAndPrometheus(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	versionInfo := &types.VersionInfo{Release: "GitRef", Sha: "GitSha", APIVersions: []types.APIVersion{types.APIV2}}
	gin.SetMode(gin.ReleaseMode)
	router := CreateRouterAndPrometheus(versionInfo, config, newMockedCoordinator)
	assert.NotNil(t, router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8; escaping=underscores", w.Result().Header.Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "migrator_gin_info")
}

// /health
func TestHealthCheck(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, newMockedCoordinator)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `"status":"UP"`)
}

func TestHealthCheckServiceUnavailable(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, newMockedCoordinatorHealthCheckError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `"status":"DOWN"`)
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

func TestConfigRouteV2(t *testing.T) {
	configObj, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(configObj, nil)

	w := httptest.NewRecorder()
	req, _ := newTestRequestV2("GET", "/config", nil)
	router.ServeHTTP(w, req)

	// gin uses custom YAML serialization
	actual, _ := config.FromBytes(w.Body.Bytes())

	// returns config
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/yaml; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Equal(t, configObj.String(), actual.String())
}

func TestGraphQLSchema(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, newMockedCoordinator)

	w := httptest.NewRecorder()
	req, _ := newTestRequestV2("GET", "/schema", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Result().Header.Get("Content-Type"))
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
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
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
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Equal(t, `{"errors":[{"message":"Invalid request, please see documentation for valid JSON payload"}]}`, strings.TrimSpace(w.Body.String()))
}

func TestPanicHandlerGlobal(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouterInDebug(config, newMockedErrorCoordinator(0))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Equal(t, `{"errors":[{"message":"Mocked Coordinator: threshold 0 reached"}]}`, strings.TrimSpace(w.Body.String()))
}

func TestPanicHandlerGraphql(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	router := testSetupRouter(config, newMockedErrorCoordinator(0))

	w := httptest.NewRecorder()
	req, _ := newTestRequestV2("POST", "/service", strings.NewReader(`
    {
      "query": "query SourceMigration($file: String!) { sourceMigration(file: $file) { name, migrationType, sourceDir, file } }",
      "operationName": "SourceMigration",
      "variables": { "file": "source/201602220001.sql" }
    }
  `))
	router.ServeHTTP(w, req)

	body := w.Body.String()

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Result().Header.Get("Content-Type"))
	assert.Contains(t, body, "panic occurred: Mocked Coordinator: threshold 0 reached")
}
