package feedgen

import (
	"testing"
	"time"

	"goalgotrade/pkg/core"
)

func TestYahooSimple(t *testing.T) {
	rtn := NewYahooBarFeedGenerator("GLD", core.DAY)
	if rtn == nil {
		t.Error("Expected non-nil return")
	}
}

func TestYahooSimple2(t *testing.T) {
	gen := NewYahooBarFeedGenerator("GLD", core.DAY)
	disp := core.NewDispatcher()
	feed := core.NewGenericDataFeed(gen, 100)
	disp.AddSubject(feed)

	go disp.Run()

	time.Sleep(time.Second * 2)
	disp.Stop()
}
