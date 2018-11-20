// These are integration tests.
// The following tests must be working in order to get this one working:
// * config_test.go
// * migrations_test.go
// DB & Disk operations are mocked using xcli_mocks.go

package server

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	configFile = "../test/migrator.yaml"
)

func TestRegisterHandlers(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)
	registerHandlers(config, nil, nil)
}

func TestServerDefaultHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/qwq", nil)

	w := httptest.NewRecorder()
	defaultHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServerConfig(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", "http://example.com/config", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(configHandler, config, nil, nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
}

func TestServerDBTenants(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", "http://example.com/dbTenants", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(dbTenantsHandler, config, createMockedConnector, nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `["a","b","c"]`, strings.TrimSpace(w.Body.String()))
}

func TestServerDBMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", "http://example.com/dbMigrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(dbMigrationsHandler, config, createMockedConnector, nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Schema":"source","Created":"2016-02-22T16:41:01.000000123Z"}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerDiskMigrations(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", "http://example.com/dbMigrations", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(diskMigrationsHandler, config, nil, createMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220000.sql","SourceDir":"source","File":"source/201602220000.sql","MigrationType":1,"Contents":"select abc"},{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":1,"Contents":"select def"}]`, strings.TrimSpace(w.Body.String()))
}

func TestServerApplyError(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", "http://example.com/apply", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(applyHandler, config, nil, nil)
	handler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestServerApplyOK(t *testing.T) {
	config, err := config.FromFile(configFile)
	assert.Nil(t, err)

	req, _ := http.NewRequest("POST", "http://example.com/apply", nil)

	w := httptest.NewRecorder()
	handler := makeHandler(applyHandler, config, createMockedConnector, createMockedDiskLoader)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.HeaderMap["Content-Type"][0])
	assert.Equal(t, `[{"Name":"201602220001.sql","SourceDir":"source","File":"source/201602220001.sql","MigrationType":1,"Contents":"select def"}]`, strings.TrimSpace(w.Body.String()))
}
