package notifier

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	lg "goalgotrade/logger"
	"strconv"

	"gopkg.in/gomail.v2"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
)

// Notifier ...
type Notifier interface {
	SendMessage(from, to, title, body string) (string, error)
	Poke(from, to, message string) (string, error)
}

type smtpEmailNotifier struct {
	host        string
	port        int
	username    string
	password    string
	defaultFrom string
}

// NewSMTPEmailNotifier ...
func NewSMTPEmailNotifier(host, portStr, username, password, defaultFrom string) Notifier {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("invalid port")
	}
	return &smtpEmailNotifier{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		defaultFrom: defaultFrom,
	}
}

// SendMessage ...
func (s *smtpEmailNotifier) SendMessage(from, to, title, body string) (string, error) {
	if from == "" {
		from = s.defaultFrom
	}
	if to == "" {
		return "", fmt.Errorf("invalid target address")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", body)
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		lg.Logger.Error("error sending email", zap.Error(err))
		return "", err
	}
	return "", nil
}

// Poke ...
func (s *smtpEmailNotifier) Poke(from, to, message string) (string, error) {
	return s.SendMessage(from, to, "Poke!👉🏼", message)
}

type twilioNotifier struct {
	accountSid          string
	authToken           string
	messagingServiceSid string
	defaultTo           string
	client              *twilio.RestClient
}

// SendMessage ...
func (t *twilioNotifier) SendMessage(from, to, title, body string) (string, error) {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	if from != "" {
		params.SetFrom(from)
	} else {
		params.SetMessagingServiceSid(t.messagingServiceSid)
	}
	params.SetBody(body)

	resp, err := t.client.ApiV2010.CreateMessage(params)
	if err != nil {
		return "", err
	} else {
		response, err := json.Marshal(*resp)
		if err != nil {
			return string(response), err
		}
		return string(response), nil
	}
}

// Poke ...
func (t *twilioNotifier) Poke(from, to, message string) (string, error) {
	twiml := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Response><Pause length=\"2\" />"+
		"<Say>%s</Say><Pause length=\"5\" /><Hangup /></Response>", message)
	params := &openapi.CreateCallParams{}
	params.SetTo(to)
	params.SetFrom(from)
	params.SetTwiml(twiml)

	resp, err := t.client.ApiV2010.CreateCall(params)
	if err != nil {
		return "", err
	} else {
		response, err := json.Marshal(*resp)
		if err != nil {
			return string(response), err
		}
		return string(response), nil
	}
}

// NewTwilioNotifier ...
func NewTwilioNotifier(accountSid, authToken, messagingServiceSid, defaultTo string) Notifier {
	return &twilioNotifier{
		accountSid:          accountSid,
		authToken:           authToken,
		messagingServiceSid: messagingServiceSid,
		defaultTo:           defaultTo,
		client: twilio.NewRestClientWithParams(
			twilio.RestClientParams{
				Username: accountSid,
				Password: authToken,
			}),
	}
}
