package test

import (
	"flag"
	"goalgotrade/broker"
	"goalgotrade/consts/frequency"
	"goalgotrade/feed"
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
	freqList := []frequency.Frequency{frequency.REALTIME, frequency.MINUTE}

	tvf := feed.NewFakeFetcherProvider()
	bbf := feed.NewBaseBarFetcher(tvf, 3*time.Second)

	if err := bbf.RegisterInstrument(symbol, freqList); err != nil {
		t.Error(err)
		return
	}

	lbf := feed.NewLiveBarFeed(bbf, 100)
	if lbf == nil {
		t.Error("cannot create live bar feed")
	}

	if err := bbf.Start(); err != nil {
		t.Error(err)
		return
	}

	b := broker.NewBaseBroker(lbf)
	s := strategy.NewBaseStrategy(lbf, b)

	go func() {
		var err error

		timer := time.NewTimer(20 * time.Second)

		select {
		case <-timer.C:
		}

		t.Log("terminating...")

		if bbf.Stop() != nil {
			t.Error(err)
			return
		}

		if s.Stop() != nil {
			t.Error(err)
			return
		}
	}()

	err := s.Run(s)
	if err != nil {
		panic(err)
	}
}
