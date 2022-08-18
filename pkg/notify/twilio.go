package notify

import (
	"encoding/json"
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
)

type twilioNotifier struct {
	cfg     *config.Config
	sid     string
	tok     string
	to      []string
	from    string
	subject string
	message string
}

// Send implements Notifier
func (t *twilioNotifier) Send() error {
	if len(t.to) == 0 {
		return fmt.Errorf("no recipients for twilio notification")
	}

	if t.sid == "" || t.tok == "" {
		return fmt.Errorf("twilio SID and token must be set")
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: t.sid,
		Password: t.tok,
	})

	for _, recipient := range t.to {
		params := &openapi.CreateMessageParams{}
		params.SetTo(recipient)
		params.SetFrom(t.from)
		params.SetBody(fmt.Sprintf("%s: %s", t.subject, t.message))

		resp, err := client.Api.CreateMessage(params)
		if err != nil {
			logger.Logger.Error("error sending twilio notification", zap.Error(err))
			return err
		} else {
			response, _ := json.Marshal(*resp)
			logger.Logger.Debug("twilio notification sent", zap.String("response", string(response)))
		}
	}

	return nil
}

// SetContent implements Notifier
func (t *twilioNotifier) SetContent(c string) error {
	t.message = c
	return nil
}

// SetSubject implements Notifier
func (t *twilioNotifier) SetSubject(s string) error {
	t.subject = s
	return nil
}

// SetRecipient implements Notifier
func (t *twilioNotifier) SetRecipients(r []string) error {
	t.to = r
	return nil
}

func NewTwilioNotifier(cfg *config.Config) Notifier {
	return &twilioNotifier{
		cfg:  cfg,
		sid:  cfg.Notification.Twilio.SID,
		tok:  cfg.Notification.Twilio.Token,
		from: cfg.Notification.Twilio.From,
		to:   cfg.Notification.Twilio.To,
	}
}
