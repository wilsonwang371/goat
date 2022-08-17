package notify

import (
	"crypto/tls"

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

// Send implements Notifier
func (e *emailNotifier) Send() error {
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
	e.content = c
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
