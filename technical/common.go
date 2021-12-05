package technical

import (
	"goalgotrade/core"
	"goalgotrade/dataseries"
	"sync"
	"time"
)

type EventWindow interface {
	OnNewValue(dateTime *time.Time, value interface{})
	WindowSize() int
	Values() []interface{}
	IsWindowFull() bool
	Value() interface{}
}

type baseEventWindow struct {
	mu         sync.Mutex
	values     []interface{}
	windowSize int
	skipNone   bool
}

func NewEventWindow(windowSize int, skipNone bool) EventWindow {
	return newEventWindow(windowSize, skipNone)
}

func newEventWindow(windowSize int, skipNone bool) *baseEventWindow {
	res := &baseEventWindow{
		windowSize: windowSize,
		skipNone:   skipNone,
	}
	return res
}

func (e *baseEventWindow) OnNewValue(dateTime *time.Time, value interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if value != nil || !e.skipNone {
		e.values = append(e.values, value)
		if len(e.values) > e.windowSize {
			e.values = e.values[len(e.values)-e.windowSize:]
		}
	}
}

func (e *baseEventWindow) WindowSize() int {
	return e.windowSize
}

func (e *baseEventWindow) Values() []interface{} {
	return e.values
}

func (e *baseEventWindow) IsWindowFull() bool {
	return len(e.values) == e.windowSize
}

func (e *baseEventWindow) Value() interface{} {
	panic("not implemented")
}

type EventBasedFilter interface {
	DataSeries() dataseries.SequenceDataSeries
	EventWindow() EventWindow
}

type eventBasedFilter struct {
	dataseries.SequenceDataSeries
	mu          sync.Mutex
	dataSeries  dataseries.SequenceDataSeries
	eventWindow EventWindow
}

func NewEventBasedFilter(dataSeries dataseries.SequenceDataSeries, eventWindow EventWindow, maxLen int) EventBasedFilter {
	return newEventBasedFilter(dataSeries, eventWindow, maxLen)
}

func newEventBasedFilter(dataSeries dataseries.SequenceDataSeries, eventWindow EventWindow, maxLen int) *eventBasedFilter {
	res := &eventBasedFilter{
		SequenceDataSeries: dataseries.NewSequenceDataSeries(maxLen),
		dataSeries:         dataSeries,
		eventWindow:        eventWindow,
	}
	res.dataSeries.NewValueChannel().Subscribe(func(event core.Event) error {
		var d dataseries.SequenceDataSeries
		var t *time.Time
		var v interface{}

		if tmp, ok := event.Get("dataseries"); ok {
			d = tmp.(dataseries.SequenceDataSeries)
		}
		if tmp, ok := event.Get("time"); ok {
			t = tmp.(*time.Time)
		}
		if tmp, ok := event.Get("value"); ok {
			v = tmp
		}
		return res.onNewValue(d, t, v)
	})
	return res
}

func (f *eventBasedFilter) onNewValue(dataSeries dataseries.SequenceDataSeries, dateTime *time.Time, value interface{}) error {
	f.eventWindow.OnNewValue(dateTime, value)
	newValue := f.eventWindow.Value()
	return f.AppendWithDateTime(dateTime, newValue)
}

func (f *eventBasedFilter) DataSeries() dataseries.SequenceDataSeries {
	return f.dataSeries
}

func (f *eventBasedFilter) EventWindow() EventWindow {
	return f.eventWindow
}
