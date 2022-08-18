package feedgen

import (
	"os"
	"testing"
	"time"

	"goat/pkg/core"
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

	go gen.Run()
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

	go gen.Run()
	go disp.Run()

	time.Sleep(time.Second * 5)
	disp.Stop()
}

func TestFx678StringGen(t *testing.T) {
	count := 0

	for {
		if bar, err := getABar("XAU"); err != nil {
			t.Error(err)
			count++
			time.Sleep(time.Second * 5)
		} else {
			t.Log(bar)
			return
		}

		if count > 10 {
			t.Error("failed to get bar")
			return
		}
	}
}
