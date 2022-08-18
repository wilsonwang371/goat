package notify

import (
	"os"
	"testing"

	"goat/pkg/config"
)

const (
	GoatTwilioAccountSid = "GOAT_TWILIO_ACCOUNT_SID"
	GoatTwilioAuthToken  = "GOAT_TWILIO_AUTH_TOKEN"
	GoatTwilioFrom       = "GOAT_TWILIO_FROM"
	GoatTwilioTo         = "GOAT_TWILIO_TO"
)

func TestTwilioSimple(t *testing.T) {
	cfg := &config.Config{}

	if sid, exists := os.LookupEnv(GoatTwilioAccountSid); exists {
		cfg.Notification.Twilio.SID = sid
	} else {
		t.Skip("environment variable " + GoatTwilioAccountSid + " is not set")
	}

	if token, exists := os.LookupEnv(GoatTwilioAuthToken); exists {
		cfg.Notification.Twilio.Token = token
	} else {
		t.Skip("environment variable " + GoatTwilioAuthToken + " is not set")
	}

	if from, exists := os.LookupEnv(GoatTwilioFrom); exists {
		cfg.Notification.Twilio.From = from
	} else {
		t.Skip("environment variable " + GoatTwilioFrom + " is not set")
	}

	if to, exists := os.LookupEnv(GoatTwilioTo); exists {
		cfg.Notification.Twilio.To = []string{to}
	} else {
		t.Skip("environment variable " + GoatTwilioTo + " is not set")
	}

	cli := NewTwilioNotifier(cfg)
	cli.SetContent("this is a test")
	cli.SetSubject("subject")
	err := cli.Send()
	if err != nil {
		t.Error(err)
	}
}

func TestTwilioFailure(t *testing.T) {
	cfg := &config.Config{}
	cli := NewTwilioNotifier(cfg)
	cli.SetContent("this is a test")
	cli.SetSubject("subject")
	err := cli.Send()
	if err != nil {
		return
	}
	t.Error("expected error, got nil")
}
