package test

import (
	"goalgotrade/feed/barfeed"
	"goalgotrade/broker"
	"goalgotrade/strategy"
	"testing"

	"github.com/go-gota/gota/series"
)

func TestStrategyBasics(t *testing.T) {
	f := barfeed.NewBaseBarFeed(series.Float, 100)
	b := broker.NewBroker(f)
	s := strategy.NewBaseStrategy(f, b)
	ch, err := s.Run()
	if err != nil {
		panic(err)
	}
	<-ch
}
