package notifications

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateNoopWhenSlackWebHookNotDefined(t *testing.T) {
	config := config.Config{}
	notifier := CreateNotifier(&config)
	result, err := notifier.Notify("abc")

	assert.Equal(t, "noop", result)
	assert.Nil(t, err)
}
