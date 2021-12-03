package technical_old

import (
	"goalgotrade/common"
	"goalgotrade/dataseries"
	"sync"
	"time"
)

type EventWindow struct {
	Self       interface{}
	mu         sync.Mutex
	values     []interface{}
	windowSize int
	skipNone   bool
}

func NewEventWindow(windowSize int, skipNone bool) *EventWindow {
	res := &EventWindow{
		windowSize: windowSize,
		skipNone:   skipNone,
	}
	res.Self = res
	return res
}

func (e *EventWindow) OnNewValue(dateTime *time.Time, value interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if value != nil || !e.skipNone {
		e.values = append(e.values, value)
		if len(e.values) > e.windowSize {
			e.values = e.values[len(e.values)-e.windowSize:]
		}
	}
}

func (e *EventWindow) GetWindowSize() int {
	return e.windowSize
}

func (e *EventWindow) GetValues() []interface{} {
	return e.values
}

func (e *EventWindow) IsWindowFull() bool {
	return len(e.values) == e.windowSize
}

func (e *EventWindow) GetCurrentValue() interface{} {
	panic("not implemented")
}

type EventBasedFilter struct {
	dataseries.SequenceDataSeries
	Self        interface{}
	mu          sync.Mutex
	dataSeries  common_old.SequenceDataSeries
	eventWindow *EventWindow
}

func NewEventBasedFilter(dataSeries common_old.SequenceDataSeries, eventWindow *EventWindow, maxLen int) *EventBasedFilter {
	res := &EventBasedFilter{
		SequenceDataSeries: *dataseries.NewSequenceDataSeries(maxLen),
		dataSeries:         dataSeries,
		eventWindow:        eventWindow,
	}
	res.dataSeries.GetNewValueEvent().Subscribe(func(args ...interface{}) error {
		d := args[0].(common_old.SequenceDataSeries)
		t := args[1].(*time.Time)
		v := args[2]
		return res.onNewValue(d, t, v)
	})
	res.Self = res
	return res
}

func (f *EventBasedFilter) onNewValue(dataSeries common_old.SequenceDataSeries, dateTime *time.Time, value interface{}) error {
	f.eventWindow.OnNewValue(dateTime, value)
	newValue := f.eventWindow.GetCurrentValue()
	return f.AppendWithDateTime(dateTime, newValue)
}

func (f *EventBasedFilter) GetDataSeries() common_old.SequenceDataSeries {
	return f.dataSeries
}

func (f *EventBasedFilter) GetEventWindow() *EventWindow {
	return f.eventWindow
}
