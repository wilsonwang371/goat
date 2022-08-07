package core

import (
	"sync"
	"testing"
	"time"
)

func TestSimpleStrategy(t *testing.T) {
	gen := NewBarFeedGenerator(
		[]Frequency{REALTIME, DAY, HOUR},
		100)
	feed := NewGenericDataFeed(gen, 100)
	sel := NewSimpleStrategyEventListener()
	broker := NewDummyBroker(feed)
	strategy := NewStrategyController(sel, broker, feed)

	wg := &sync.WaitGroup{}

	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
			"b": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
		},
		REALTIME)
	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, DAY, time.Now()),
			"b": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, DAY, time.Now()),
		},
		DAY)
	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
			"b": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
		},
		REALTIME)
	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, HOUR, time.Now()),
			"b": NewSimpleBar(1.0, 2.0, 3.0, 1.2, 100, HOUR, time.Now()),
		},
		HOUR)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		strategy.Run()
	}(wg)

	time.Sleep(time.Second * 2)
	gen.Finish()

	wg.Wait()
}
