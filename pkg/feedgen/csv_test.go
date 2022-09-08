package feedgen

import (
	"context"
	"testing"
	"time"

	"goat/pkg/config"
	"goat/pkg/core"
)

func TestCSVSimple(t *testing.T) {
	gen := NewCSVBarFeedGenerator(
		"../../samples/data/DBC-2007-yahoofinance.csv", "Symbol",
		core.UNKNOWN)
	disp := core.NewDispatcher(context.TODO())
	feed := core.NewGenericDataFeed(context.TODO(), &config.Config{}, gen, nil, 100, "")
	disp.AddSubject(feed)

	go disp.Run()

	time.Sleep(time.Second * 2)
	disp.Stop()
}
