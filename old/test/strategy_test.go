package test

import (
	"goalgotrade/broker"
	"goalgotrade/consts/frequency"
	"goalgotrade/feed"
	"goalgotrade/strategy"
	"testing"

	"github.com/go-gota/gota/series"
)

func TestStrategyBasics(t *testing.T) {
	freqList := []frequency.Frequency{frequency.DAY, frequency.MINUTE}

	f := feed.NewBaseBarFeed(freqList, series.Float, 100)
	b := broker.NewBaseBroker(f)
	s := strategy.NewBaseStrategy(f, b)

	err := s.Run(s)
	if err != nil {
		panic(err)
	}
}
