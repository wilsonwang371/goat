package test_old

import (
	"goalgotrade/broker"
	"goalgotrade/common"
	"goalgotrade/feed/barfeed"
	"goalgotrade/strategy"
	"testing"

	"github.com/go-gota/gota/series"
)

func TestStrategyBasics(t *testing.T) {
	freqList := []common_old.Frequency{common_old.Frequency_DAY, common_old.Frequency_MINUTE}

	f := barfeed.NewBaseBarFeed(freqList, series.Float, 100)
	b := broker.NewBroker(f)
	s := strategy.NewBaseStrategy(f, b)

	ch, err := s.Run()
	if err != nil {
		panic(err)
	}

	<-ch
}
