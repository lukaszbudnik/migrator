package notifications

import (
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateSlackNotifierWhenSlackWebHookDefined(t *testing.T) {
	config := config.Config{}
	config.SlackWebHook = "https://slack.com/api/api.test"
	notifier := CreateNotifier(&config)
	result, err := notifier.Notify("abc")

	assert.NotNil(t, result)
	assert.Nil(t, err)
}

func TestSlackNotifierURLError(t *testing.T) {
	config := config.Config{}
	config.SlackWebHook = "xczxcvv"
	notifier := CreateNotifier(&config)
	result, err := notifier.Notify("abc")

	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}
