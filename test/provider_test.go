package test

import (
	"flag"
	"fmt"
	"goalgotrade/broker"
	"goalgotrade/common"
	"goalgotrade/feed/barfeed"
	"goalgotrade/feed/barfeed/fetcher"
	"goalgotrade/strategy"
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
	freqList := []common.Frequency{common.Frequency_REALTIME}

	if username == "" || password == "" {
		t.Skip("username and/or password is empty")
	}
	tvf := fetcher.NewTradingViewFetcher(username, password)

	if err := tvf.RegisterInstrument(symbol, freqList); err != nil {
		t.Error(err)
		return
	}

	lbf := barfeed.NewLiveBarFeed(tvf, 100)
	if lbf == nil {
		t.Error("cannot create live bar feed")
	}

	if err := tvf.Start(); err != nil {
		t.Error(err)
		return
	}

	b := broker.NewBroker(lbf)
	s := strategy.NewBaseStrategy(lbf, b)

	ch, err := s.Run()
	if err != nil {
		panic(err)
	}

	timer := time.NewTimer(120 * time.Second)

	select {
	case <-ch:
		t.Log("data from strategy done channel")
	case <-timer.C:
		t.Error("timeout waiting for data")
	case bars := <-tvf.PendingBarsC():
		t.Log(fmt.Sprintf("got bars: %v", bars))
	case err := <-tvf.ErrorC():
		t.Error(err)
	}

	if err := tvf.Stop(); err != nil {
		t.Error(err)
		return
	}
}
