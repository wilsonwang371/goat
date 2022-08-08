package feedgen

import (
	"goalgotrade/pkg/core"
	"os"
	"testing"
	"time"
)

func TestTradingViewSimple(t *testing.T) {
	user := os.Getenv("TRADINGVIEW_USER")
	pass := os.Getenv("TRADINGVIEW_PASS")
	if user == "" || pass == "" {
		t.Skip("TRADINGVIEW_USER and TRADINGVIEW_PASS must be set")
	}
	gen := NewLiveBarFeedGenerator(
		NewTradingViewDataProvider(user, pass),
		"XAUUSD",
		[]core.Frequency{core.REALTIME, core.DAY},
		100)
	disp := core.NewDispatcher()
	feed := core.NewGenericDataFeed(gen, 100)
	disp.AddSubject(feed)

	go gen.(*LiveBarFeedGenerator).Run()
	go disp.Run()

	time.Sleep(time.Second * 5)
	disp.Stop()
}

func TestFakeSimple(t *testing.T) {
	gen := NewLiveBarFeedGenerator(
		NewFakeDataProvider(),
		"XAUUSD",
		[]core.Frequency{core.REALTIME, core.DAY},
		100)
	disp := core.NewDispatcher()
	feed := core.NewGenericDataFeed(gen, 100)
	disp.AddSubject(feed)

	go gen.(*LiveBarFeedGenerator).Run()
	go disp.Run()

	time.Sleep(time.Second * 5)
	disp.Stop()
}
