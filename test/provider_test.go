package test

import (
	"flag"
	"goalgotrade/providers"
	"testing"
)

var username, password string

func init() {
	flag.StringVar(&username, "username", "", "TradingView Username")
	flag.StringVar(&password, "password", "", "TradingView Password")
}

// pass "-username=<user> -password=<pass>" as arguments to test
func TestTradingViewAuth(t *testing.T) {
	if username == "" || password == "" {
		t.Skip("username and/or password is empty")
	}
	result, err := providers.GetAuthToken(username, password)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

func TestWSConnection(t *testing.T) {
	providers.TradingViewConnect()
}
