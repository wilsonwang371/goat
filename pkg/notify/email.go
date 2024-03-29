package notify

import (
	"crypto/tls"
	"fmt"
	"strings"

	"goat/pkg/config"

	"gopkg.in/gomail.v2"
)

type emailNotifier struct {
	cfg *config.Config

	subject    string
	from       string
	recipients []string
	content    string
}

// FeatureFlags implements Notifier
func (*emailNotifier) FeatureFlags() uint64 {
	return config.NotifyIsEmailFlag
}

// Level implements Notifier
func (*emailNotifier) Level() int {
	return config.InfoLevel
}

// Send implements Notifier
func (e *emailNotifier) Send() error {
	if len(e.recipients) == 0 {
		return fmt.Errorf("no recipients for email notification")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", e.recipients...)
	m.SetHeader("Subject", e.subject)
	m.SetBody("text/html", e.content)

	d := gomail.NewDialer(e.cfg.Notification.Email.Host,
		e.cfg.Notification.Email.Port,
		e.cfg.Notification.Email.User,
		e.cfg.Notification.Email.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// SetContent implements Notifier
func (e *emailNotifier) SetContent(c string) error {
	if strings.Contains(c, "<html>") {
		e.content = c
	} else {
		newC := strings.Replace(c, "\n", "<br>", -1)
		e.content = fmt.Sprintf("<html><body>%s</body></html>", newC)
	}
	return nil
}

// SetSubject implements Notifier
func (e *emailNotifier) SetSubject(s string) error {
	e.subject = s
	return nil
}

// SetRecipient implements Notifier
func (e *emailNotifier) SetRecipients(r []string) error {
	e.recipients = r
	return nil
}

func NewEmailNotifier(cfg *config.Config) Notifier {
	return &emailNotifier{
		cfg:        cfg,
		from:       cfg.Notification.Email.From,
		recipients: cfg.Notification.Email.To,
	}
}
