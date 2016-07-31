package notifications

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSlackNotifierWhenSlackWebHookDefined(t *testing.T) {
	config := config.Config{}
	config.SlackWebHook = "https://slack.com/api/api.test"
	notifier := CreateNotifier(&config)
	result, err := notifier.Notify("abc")

	assert.NotNil(t, result)
	assert.Nil(t, err)
}
