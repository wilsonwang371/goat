package test

import (
	"flag"
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
	flag.StringVar(&symbol, "symbol", "BINANCE:BTCUSDT", "TradingView Symbol")
}

// pass "-username=<user> -password=<pass>" as arguments to test
func TestTradingView(t *testing.T) {
	freqList := []common.Frequency{common.Frequency_REALTIME, common.Frequency_MINUTE}

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

	go func() {
		var err error

		timer := time.NewTimer(120 * time.Second)

		select {
		case <-timer.C:
		}

		t.Log("terminating...")

		if tvf.Stop() != nil {
			t.Error(err)
			return
		}

		if s.Stop() != nil {
			t.Error(err)
			return
		}
	}()

	err := s.Run()
	if err != nil {
		panic(err)
	}
}
