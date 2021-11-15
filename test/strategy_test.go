package test

import (
	"goalgotrade/barfeed"
	"goalgotrade/broker"
	"goalgotrade/strategy"
	"testing"
)

func TestStrategyBasics(t *testing.T) {
	f := barfeed.NewBarFeed()
	b := broker.NewBroker()
	s := strategy.NewBaseStrategy(f, b)
	s.Run()
}
