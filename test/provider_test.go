package test

import (
	"flag"
	"fmt"
	"goalgotrade/common"
	"goalgotrade/feed/barfeed/fetcher"
	"testing"
	"time"
)

var username, password, symbol string

func init() {
	flag.StringVar(&username, "username", "", "TradingView Username")
	flag.StringVar(&password, "password", "", "TradingView Password")
	flag.StringVar(&symbol, "symbol", "FOREXCOM:XAUUSD", "TradingView Symbol")
}

// pass "-username=<user> -password=<pass>" as arguments to test
func TestTradingView(t *testing.T) {
	if username == "" || password == "" {
		t.Skip("username and/or password is empty")
	}
	tv := fetcher.NewTradingViewFetcher(username, password)

	if err := tv.RegisterInstrument(symbol, []common.Frequency{common.Frequency_REALTIME}); err != nil {
		t.Error(err)
		return
	}

	if err := tv.Start(); err != nil {
		t.Error(err)
		return
	}

	timer := time.NewTimer(60 * time.Second)

	select {
	case <-timer.C:
		t.Error("timeout waiting for data")
	case bars := <-tv.PendingBarsC():
		t.Log(fmt.Sprintf("got bars: %v", bars))
	case err := <-tv.ErrorC():
		t.Error(err)
	}

	if err := tv.Stop(); err != nil {
		t.Error(err)
		return
	}
}
