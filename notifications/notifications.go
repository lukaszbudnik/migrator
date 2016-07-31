package notifications

import (
	"github.com/lukaszbudnik/migrator/config"
)

// Notifier interface abstracts all notifications performed by migrator
type Notifier interface {
	Notify(string) (string, error)
}

// BaseNotifier type is a base struct embedded by all implementations of Notifier interface
type BaseNotifier struct {
	Config *config.Config
}

type noopNotifier struct {
	BaseNotifier
}

func (sn *noopNotifier) Notify(text string) (string, error) {
	return "noop", nil
}

// CreateNotifier creates Notifier object based on config passed
func CreateNotifier(config *config.Config) Notifier {
	if len(config.SlackWebHook) > 0 {
		return &slackNotifier{BaseNotifier{config}}
	}
	return &noopNotifier{BaseNotifier{config}}
}
