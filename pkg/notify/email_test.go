package notify

import (
	"os"
	"strconv"
	"testing"

	"goat/pkg/config"
)

/*

you need to set the following environment variables to run this test:

GOAT_EMAIL_HOST=
GOAT_EMAIL_PORT=
GOAT_EMAIL_USER=
GOAT_EMAIL_PASSWORD=
GOAT_EMAIL_FROM=
GOAT_EMAIL_TO=

*/

const (
	GoatEmailHost = "GOAT_EMAIL_HOST"
	GoatEmailPort = "GOAT_EMAIL_PORT"
	GoatEmailUser = "GOAT_EMAIL_USER"
	GoatEmailPass = "GOAT_EMAIL_PASSWORD"
	GoatEmailFrom = "GOAT_EMAIL_FROM"
	GoatEmailTo   = "GOAT_EMAIL_TO"
)

func TestEmailSimple(t *testing.T) {
	cfg := &config.Config{}

	if host, exists := os.LookupEnv(GoatEmailHost); exists {
		cfg.Notification.Email.Host = host
	} else {
		t.Skip("environment variable " + GoatEmailHost + " is not set")
	}

	if port, exists := os.LookupEnv(GoatEmailPort); exists {
		if tmp, err := strconv.Atoi(port); err == nil {
			cfg.Notification.Email.Port = tmp
		} else {
			t.Skip("environment variable " + GoatEmailPort + " is not a number")
		}
	} else {
		t.Skip("environment variable " + GoatEmailPort + " is not set")
	}

	if from, exists := os.LookupEnv(GoatEmailFrom); exists {
		cfg.Notification.Email.From = from
	} else {
		t.Skip("environment variable " + GoatEmailFrom + " is not set")
	}

	if to, exists := os.LookupEnv(GoatEmailTo); exists {
		cfg.Notification.Email.To = []string{to}
	} else {
		t.Skip("environment variable " + GoatEmailTo + " is not set")
	}

	if user, exists := os.LookupEnv(GoatEmailUser); exists {
		cfg.Notification.Email.User = user
	} else {
		t.Skip("environment variable " + GoatEmailUser + " is not set")
	}

	if password, exists := os.LookupEnv(GoatEmailPass); exists {
		cfg.Notification.Email.Password = password
	} else {
		t.Skip("environment variable " + GoatEmailPass + " is not set")
	}

	n := NewEmailNotifier(cfg)
	n.SetSubject("Test Email")
	n.SetContent("This is a test email")
	if err := n.Send(); err != nil {
		t.Error(err)
	}
}
