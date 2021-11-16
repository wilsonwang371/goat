package test

import (
	"goalgotrade/barfeed"
	"goalgotrade/broker"
	"goalgotrade/strategy"
	"testing"
)

func TestStrategyBasics(t *testing.T) {
	f := barfeed.NewBaseBarFeed(100)
	b := broker.NewBroker(f)
	s := strategy.NewBaseStrategy(f, b)
	ch, err := s.Run()
	if err != nil {
		panic(err)
	}
	<-ch
}
