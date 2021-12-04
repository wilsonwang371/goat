package notifier

import (
	"encoding/json"
	"fmt"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// Notifier ...
type Notifier interface {
	SendMessage(from, to, title, body string) (string, error)
	Poke(from, to, message string) (string, error)
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
