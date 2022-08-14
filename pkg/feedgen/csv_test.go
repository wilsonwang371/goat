package feedgen

import (
	"goalgotrade/pkg/core"
	"testing"
	"time"
)

func TestCSVSimple(t *testing.T) {
	gen := NewCSVBarFeedGenerator(
		"../../samples/data/DBC-2007-yahoofinance.csv", "Symbol",
		core.UNKNOWN)
	disp := core.NewDispatcher()
	feed := core.NewGenericDataFeed(gen, 100)
	disp.AddSubject(feed)

	go disp.Run()

	time.Sleep(time.Second * 2)
	disp.Stop()
}
