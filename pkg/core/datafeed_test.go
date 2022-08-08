package core

import (
	"testing"
	"time"
)

func TestSimpleDataFeedGenerator(t *testing.T) {
	gen := NewBarFeedGenerator(
		[]Frequency{REALTIME, DAY},
		100)
	disp := NewDispatcher()
	feed := NewGenericDataFeed(gen, 100)
	disp.AddSubject(feed)

	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
			"b": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
		},
		REALTIME)
	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, DAY, time.Now()),
			"b": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, DAY, time.Now()),
		},
		DAY)

	gen.Finish()

	go disp.Run()

	time.Sleep(time.Second * 2)
	disp.Stop()
}

func TestSimpleDataFeedGenerator2(t *testing.T) {
	gen := NewBarFeedGenerator(
		[]Frequency{REALTIME, DAY},
		100)
	disp := NewDispatcher()
	feed := NewGenericDataFeed(gen, 100)
	disp.AddSubject(feed)

	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
			"b": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, REALTIME, time.Now()),
		},
		REALTIME)
	gen.AppendNewValueToBuffer(time.Now(),
		map[string]interface{}{
			"a": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, DAY, time.Now()),
			"b": NewBasicBar(1.0, 2.0, 3.0, 1.2, 100, DAY, time.Now()),
		},
		DAY)

	gen.Finish()
	go disp.Run()

	time.Sleep(time.Second * 2)
	gen.Finish()
}
