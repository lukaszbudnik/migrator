package notifications

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/lukaszbudnik/migrator/config"
)

const (
	defaultContentType = "application/json"
	textPlaceHolder    = "{text}"
)

// Notifier interface abstracts all notifications performed by migrator
type Notifier interface {
	Notify(string) (string, error)
}

// baseNotifier type is a base struct embedded by all implementations of Notifier interface
type baseNotifier struct {
	config *config.Config
}

func (bn *baseNotifier) Notify(message string) (string, error) {
	reader := bytes.NewReader([]byte(message))

	url := bn.config.WebHookURL

	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return "", err
	}
	for _, header := range bn.config.WebHookHeaders {
		pair := strings.SplitN(header, ":", 2)
		req.Header.Set(pair[0], pair[1])
	}

	// set default content type
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", defaultContentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("%s", b)
	return response, nil
}

type noopNotifier struct {
	baseNotifier
}

func (sn *noopNotifier) Notify(text string) (string, error) {
	return "noop", nil
}

// Factory is a factory method for creating Loader instance
type Factory func(context.Context, *config.Config) Notifier

// New creates Notifier object based on config passed
func New(ctx context.Context, config *config.Config) Notifier {
	// webhook URL is required
	if len(config.WebHookURL) > 0 {
		return &baseNotifier{config}
	}
	// otherwise return noop
	return &noopNotifier{baseNotifier{config}}
}
