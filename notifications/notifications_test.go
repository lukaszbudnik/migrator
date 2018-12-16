package notifications

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestNoopNotifier(t *testing.T) {
	config := config.Config{}
	notifier := NewNotifier(&config)
	result, err := notifier.Notify("abc")

	assert.Equal(t, "noop", result)
	assert.Nil(t, err)
}

func TestGetContentTypeDefault(t *testing.T) {
	config := config.Config{}
	config.WebHookTemplate = `{"text": "{text}","icon_emoji": ":white_check_mark:"}`
	config.WebHookURL = "https://webhook.com"

	notifier := &baseNotifier{&config}
	contentType := notifier.getContentType()

	assert.Equal(t, "application/json", contentType)
}

func TestGetContentTypeOverride(t *testing.T) {
	config := config.Config{}
	config.WebHookTemplate = `{"text": "{text}","icon_emoji": ":white_check_mark:"}`
	config.WebHookURL = "https://webhook.com"
	config.WebHookContentType = "application/x-www-form-urlencoded"

	notifier := &baseNotifier{&config}
	contentType := notifier.getContentType()

	assert.Equal(t, config.WebHookContentType, contentType)
}

func TestWebHookNotifier(t *testing.T) {

	var request []byte

	server := httptest.NewServer(func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			request, _ = ioutil.ReadAll(r.Body)
			w.Write([]byte(`{"result": "ok"}`))
		}
	}())

	config := config.Config{}
	config.WebHookTemplate = `{"text": "{text}","icon_emoji": ":white_check_mark:"}`
	config.WebHookURL = server.URL

	notifier := NewNotifier(&config)

	result, err := notifier.Notify("abc")

	assert.Nil(t, err)
	assert.Equal(t, `{"result": "ok"}`, result)
	assert.Equal(t, `{"text": "abc","icon_emoji": ":white_check_mark:"}`, string(request))
}

func TestWebHookURLError(t *testing.T) {
	config := config.Config{}
	config.WebHookURL = "xczxcvv"
	config.WebHookTemplate = "not imporant for this test"
	notifier := NewNotifier(&config)
	result, err := notifier.Notify("abc")

	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}
