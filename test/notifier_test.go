package test

import (
	"flag"
	"goalgotrade/notifier"
	"testing"
)

var sid, token, from, to string

var emailHost, emailPort, emailUser, emailPass, emailFrom, emailTo string

func init() {
	flag.StringVar(&sid, "twSid", "", "Account SID")
	flag.StringVar(&token, "twToken", "", "Auth Token")
	flag.StringVar(&from, "twFrom", "", "From Phone Number")
	flag.StringVar(&to, "twTo", "", "To Phone Number")

	flag.StringVar(&emailHost, "emailHost", "", "Email Host")
	flag.StringVar(&emailPort, "emailPort", "", "Email Port")
	flag.StringVar(&emailUser, "emailUser", "", "Email Username")
	flag.StringVar(&emailPass, "emailPass", "", "Email Password")
	flag.StringVar(&emailFrom, "emailFrom", "", "Email Sender")
	flag.StringVar(&emailTo, "emailTo", "", "Email Receiver")
}

// pass "-twSid=<sid> -twToken=<token> -twFrom=<from_number> -twTo=<to_number>" as arguments to test
func TestTwilioNotifier(t *testing.T) {
	if sid == "" || token == "" {
		t.Skip("twilio account not setuped yet")
	}
	n := notifier.NewTwilioNotifier(sid, token, "", to)
	result, err := n.SendMessage(from, to, "test", "test")
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
	result, err = n.Poke(from, to, "this is a test message!")
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestSMPTEmailNotifier(t *testing.T) {
	if emailHost == "" || emailPass == "" || emailUser == "" || emailPort == "" {
		t.Skip("smtp email notification not setuped yet")
	}
	n := notifier.NewSMTPEmailNotifier(emailHost, emailPort, emailUser, emailPass, emailFrom)
	_, err := n.Poke(emailFrom, emailTo, "poke test")
	if err != nil {
		t.Error(err)
	}
}
