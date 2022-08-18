package notify

import (
	"os"
	"testing"

	"goat/pkg/config"
)

const (
	GoatPushoverToken = "GOAT_PUSHOVER_TOKEN"
	GoatPushoverKey   = "GOAT_PUSHOVER_KEY"
)

func TestPushoverSimple(t *testing.T) {
	cfg := &config.Config{}

	if token, exists := os.LookupEnv(GoatPushoverToken); exists {
		cfg.Notification.Pushover.Token = token
	} else {
		t.Skip("environment variable " + GoatPushoverToken + " is not set")
	}

	if user, exists := os.LookupEnv(GoatPushoverKey); exists {
		cfg.Notification.Pushover.Keys = []string{user}
	} else {
		t.Skip("environment variable " + GoatPushoverKey + " is not set")
	}

	notifier := NewPushoverNotifier(cfg)
	notifier.SetContent("this is a test")
	notifier.SetSubject("subject")
	err := notifier.Send()
	if err != nil {
		t.Error(err)
	}
}

func TestPushoverFailure(t *testing.T) {
	cfg := &config.Config{}
	notifier := NewPushoverNotifier(cfg)
	notifier.SetContent("this is a test")
	notifier.SetSubject("subject")
	err := notifier.Send()
	if err != nil {
		return
	}
	t.Error("expected error, got nil")
}
