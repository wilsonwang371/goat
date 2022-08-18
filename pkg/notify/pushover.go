package notify

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/gregdel/pushover"
	"go.uber.org/zap"
)

type pushoverNotifier struct {
	cfg        *config.Config
	content    string
	subject    string
	recipients []string
}

// Send implements Notifier
func (p *pushoverNotifier) Send() error {
	app := pushover.New(p.cfg.Notification.Pushover.Token)

	message := pushover.NewMessage(fmt.Sprintf("%s: %s", p.subject, p.content))

	var recipients []string
	if len(p.recipients) > 0 {
		recipients = p.recipients
	} else {
		recipients = p.cfg.Notification.Pushover.Keys
	}

	if len(recipients) == 0 {
		return fmt.Errorf("no recipients for pushover notification")
	}

	for _, recipient := range recipients {
		recipient := pushover.NewRecipient(recipient)
		response, err := app.SendMessage(message, recipient)
		if err != nil {
			logger.Logger.Error("error sending pushover notification: %s", zap.Error(err))
		}
		logger.Logger.Debug("pushover notification sent: %s", zap.Any("response", response))
	}
	return nil
}

// SetContent implements Notifier
func (p *pushoverNotifier) SetContent(c string) error {
	p.content = c
	return nil
}

// SetSubject implements Notifier
func (p *pushoverNotifier) SetSubject(s string) error {
	p.subject = s
	return nil
}

// SetRecipient implements Notifier
func (p *pushoverNotifier) SetRecipients(recipients []string) error {
	p.recipients = recipients
	return nil
}

func NewPushoverNotifier(cfg *config.Config) Notifier {
	return &pushoverNotifier{
		cfg: cfg,
	}
}
