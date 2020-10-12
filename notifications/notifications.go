package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	gojsonq "github.com/thedevsaddam/gojsonq/v2"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

const (
	defaultContentType = "application/json"
)

// Notifier interface abstracts all notifications performed by migrator
type Notifier interface {
	Notify(*types.Summary) (string, error)
}

// baseNotifier type is a base struct embedded by all implementations of Notifier interface
type baseNotifier struct {
	config *config.Config
}

func (bn *baseNotifier) Notify(summary *types.Summary) (string, error) {
	summaryJSON, err := json.MarshalIndent(summary, "", "    ")
	if err != nil {
		return "", err
	}

	payload := string(summaryJSON)

	if template := bn.config.WebHookTemplate; len(template) > 0 {
		// ${summary} will be replaced with the JSON object
		if strings.Contains(template, "${summary}") {
			template = strings.Replace(template, "${summary}", strings.ReplaceAll(payload, "\"", "\\\""), -1)
		}
		// migrator also supports parsing individual fields using ${summary.field} syntax
		if strings.Contains(template, "${summary.") {
			r, _ := regexp.Compile("\\${summary.([a-zA-Z]+)}")
			matches := r.FindAllStringSubmatch(template, -1)
			for _, m := range matches {
				value := gojsonq.New().FromString(payload).Find(m[1])
				valueString := fmt.Sprintf("%v", value)
				template = strings.Replace(template, m[0], valueString, -1)
			}
		}
		payload = template
	}

	reader := bytes.NewReader([]byte(payload))
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

func (sn *noopNotifier) Notify(summary *types.Summary) (string, error) {
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
