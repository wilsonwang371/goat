package test

import (
	"flag"
	"goalgotrade/notifier"
	"testing"
)

var sid, token, from, to string

func init() {
	flag.StringVar(&sid, "sid", "", "Account SID")
	flag.StringVar(&token, "token", "", "Auth Token")
	flag.StringVar(&from, "from", "", "From Phone Number")
	flag.StringVar(&to, "to", "", "To Phone Number")
}

// pass "-sid=<sid> -token=<token> -from=<from_number> -to=<to_number>" as arguments to test
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
