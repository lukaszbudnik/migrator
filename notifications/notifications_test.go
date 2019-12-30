package notifications

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestNoopNotifier(t *testing.T) {
	config := config.Config{}
	notifier := New(context.TODO(), &config)
	result, err := notifier.Notify("abc")

	assert.Equal(t, "noop", result)
	assert.Nil(t, err)
}

func TestWebHookNotifier(t *testing.T) {

	var request []byte
	var contentType string

	server := httptest.NewServer(func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			request, _ = ioutil.ReadAll(r.Body)
			contentType = r.Header.Get("Content-Type")
			w.Write([]byte(`{"result": "ok"}`))
		}
	}())

	config := config.Config{}
	config.WebHookURL = server.URL

	notifier := New(context.TODO(), &config)

	message := `{"text": "abc","icon_emoji": ":white_check_mark:"}`
	result, err := notifier.Notify(message)

	assert.Nil(t, err)
	assert.Equal(t, `{"result": "ok"}`, result)
	assert.Equal(t, `{"text": "abc","icon_emoji": ":white_check_mark:"}`, string(request))
	assert.Equal(t, "application/json", string(contentType))
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
	config.WebHookTemplate = `{"text": "{text}","icon_emoji": ":white_check_mark:"}`
	config.WebHookURL = server.URL
	config.WebHookHeaders = []string{"Authorization: Basic QWxhZGRpbjpPcGVuU2VzYW1l", "Content-Type: application/x-yaml", "X-CustomHeader: value1,value2"}

	notifier := New(context.TODO(), &config)

	result, err := notifier.Notify("abc")

	assert.Nil(t, err)
	assert.Equal(t, `{"result": "ok"}`, result)
	assert.Equal(t, `Basic QWxhZGRpbjpPcGVuU2VzYW1l`, string(authorizationHeader))
	assert.Equal(t, `application/x-yaml`, string(contentTypeHeader))
	assert.Equal(t, `value1,value2`, string(xCustomHeader))
}

func TestWebHookURLError(t *testing.T) {
	config := config.Config{}
	config.WebHookURL = "xczxcvv"
	config.WebHookTemplate = "not imporant for this test"
	notifier := New(context.TODO(), &config)
	result, err := notifier.Notify("abc")

	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}
