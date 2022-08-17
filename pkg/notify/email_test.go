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

func TestEmailSimple(t *testing.T) {
	cfg := &config.Config{}

	if host, exists := os.LookupEnv("GOAT_EMAIL_HOST"); exists {
		cfg.Notification.Email.Host = host
	} else {
		t.Skip("GOAT_EMAIL_HOST not set")
	}

	if port, exists := os.LookupEnv("GOAT_EMAIL_PORT"); exists {
		if tmp, err := strconv.Atoi(port); err == nil {
			cfg.Notification.Email.Port = tmp
		} else {
			t.Skip("GOAT_EMAIL_PORT not a number")
		}
	} else {
		t.Skip("GOAT_EMAIL_PORT not set")
	}

	if from, exists := os.LookupEnv("GOAT_EMAIL_FROM"); exists {
		cfg.Notification.Email.From = from
	} else {
		t.Skip("GOAT_EMAIL_FROM not set")
	}

	if to, exists := os.LookupEnv("GOAT_EMAIL_TO"); exists {
		cfg.Notification.Email.To = []string{to}
	} else {
		t.Skip("GOAT_EMAIL_TO not set")
	}

	if user, exists := os.LookupEnv("GOAT_EMAIL_USER"); exists {
		cfg.Notification.Email.User = user
	} else {
		t.Skip("GOAT_EMAIL_USER not set")
	}

	if password, exists := os.LookupEnv("GOAT_EMAIL_PASSWORD"); exists {
		cfg.Notification.Email.Password = password
	} else {
		t.Skip("GOAT_EMAIL_PASSWORD not set")
	}

	n := NewEmailNotifier(cfg)
	n.SetSubject("Test Email")
	n.SetContent("This is a test email")
	if err := n.Send(); err != nil {
		t.Error(err)
	}
}
