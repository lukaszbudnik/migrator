package notifications

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/stretchr/testify/assert"
)

func TestNoopNotifier(t *testing.T) {
	config := config.Config{}
	notifier := New(context.TODO(), &config)
	result, err := notifier.Notify(&types.Summary{})

	assert.Equal(t, "noop", result)
	assert.Nil(t, err)
}

func TestWebHookNotifier(t *testing.T) {

	var contentType string
	var requestBody string

	server := httptest.NewServer(func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			request, _ := ioutil.ReadAll(r.Body)
			requestBody = string(request)
			contentType = r.Header.Get("Content-Type")
			w.Write([]byte("ok"))
		}
	}())

	config := config.Config{}
	config.WebHookURL = server.URL
	config.WebHookTemplate = `{"text": "New version created: ${summary.versionId} started at: ${summary.startedAt} and took ${summary.duration}. Migrations/scripts total: ${summary.migrationsGrandTotal}/${summary.scriptsGrandTotal}. Full results are: ${summary}"}`

	notifier := New(context.TODO(), &config)

	summary := &types.Summary{
		VersionID:            213,
		StartedAt:            graphql.Time{Time: time.Now()},
		Tenants:              123,
		Duration:             3213,
		MigrationsGrandTotal: 1024,
		ScriptsGrandTotal:    74,
	}
	result, err := notifier.Notify(summary)

	assert.Nil(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, "application/json", string(contentType))
	// make sure placeholders were replaced
	assert.NotContains(t, requestBody, "${summary")
	// explicit placeholders ${summary.property}
	assert.Contains(t, requestBody, fmt.Sprint(summary.VersionID))
	// graphql.Time fields needs to be marshalled first
	startedAt, _ := summary.StartedAt.MarshalText()
	assert.Contains(t, requestBody, string(startedAt))
	assert.Contains(t, requestBody, fmt.Sprint(summary.Duration))
	assert.Contains(t, requestBody, fmt.Sprint(summary.MigrationsGrandTotal))
	assert.Contains(t, requestBody, fmt.Sprint(summary.ScriptsGrandTotal))
	// mapped as ${summary}
	assert.Contains(t, requestBody, fmt.Sprint(summary.Tenants))
}

func TestWebHookNotifierCustomHeaders(t *testing.T) {

	var xCustomHeader string
	var authorizationHeader string
	var contentTypeHeader string

	server := httptest.NewServer(func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			xCustomHeader = r.Header.Get("X-CustomHeader")
			authorizationHeader = r.Header.Get("Authorization")
			contentTypeHeader = r.Header.Get("Content-Type")
			w.Write([]byte(`{"result": "ok"}`))
		}
	}())

	config := config.Config{}
	config.WebHookURL = server.URL
	config.WebHookHeaders = []string{"Authorization: Basic QWxhZGRpbjpPcGVuU2VzYW1l", "Content-Type: application/x-yaml", "X-CustomHeader: value1,value2"}

	notifier := New(context.TODO(), &config)

	result, err := notifier.Notify(&types.Summary{})

	assert.Nil(t, err)
	assert.Equal(t, `{"result": "ok"}`, result)
	assert.Equal(t, `Basic QWxhZGRpbjpPcGVuU2VzYW1l`, string(authorizationHeader))
	assert.Equal(t, `application/x-yaml`, string(contentTypeHeader))
	assert.Equal(t, `value1,value2`, string(xCustomHeader))
}

func TestWebHookURLError(t *testing.T) {
	config := config.Config{}
	config.WebHookURL = "xczxcvv"
	notifier := New(context.TODO(), &config)
	result, err := notifier.Notify(&types.Summary{})

	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}
