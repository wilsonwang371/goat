package common

import (
	"time"

	"github.com/go-gota/gota/series"
)

type Dispatcher interface {
	AddSubject(subject Subject) error
	GetSubjects() []Subject
	GetStartEvent() Event
	GetIdleEvent() Event
	GetCurrentDateTime() *time.Time
	Stop() error
	Run() (<-chan struct{}, error)
}

type Event interface {
	Subscribe(handler EventHandler) error
	Unsubscribe(handler EventHandler) error
	Emit(args ...interface{}) []error
}

type Subject interface {
	Start() error
	Stop() error
	Join() error
	Eof() bool
	Dispatch() (bool, error)
	PeekDateTime() *time.Time
	GetDispatchPriority() int
	SetDispatchPriority(priority int)
	OnDispatcherRegistered(dispatcher Dispatcher) error
}

type Broker interface {
	Subject
	GetOrderUpdatedEvent() Event
	NotifyOrderEvent(orderEvent *OrderEvent)
	CancelOrder(order Order) error
}

type Order interface {
	GetId() uint64
	IsActive() bool
	IsFilled() bool
	GetExecutionInfo() OrderExecutionInfo
	AddExecutionInfo(info OrderExecutionInfo) error
	GetRemaining() int
	SwitchState(newState OrderState) error
}

type Bar interface {
	SetUseAdjustedValue(useAdjusted bool) error
	GetUseAdjValue() bool
	GetDateTime() *time.Time
	Open(adjusted bool) float64
	High(adjusted bool) float64
	Low(adjusted bool) float64
	Close(adjusted bool) float64
	Volume() int
	AdjClose() float64
	Frequency() Frequency
	Price() float64
}

type Bars interface {
	GetDateTime() *time.Time
	GetInstruments() []string
	GetBar(instrument string) Bar
	GetFrequencies() []Frequency
	AddBar(instrument string, bar Bar) error
}

type Feed interface {
	Subject
	CreateDataSeries(key string, maxlen int) *series.Series
	GetNextValues() (*time.Time, Bars, Frequency, error)
	GetNextValuesAndUpdateDS() (*time.Time, Bars, Frequency, error)
	RegisterDataSeries(key string, freq Frequency) error
	GetNewValuesEvent() Event
	Reset()
	GetKeys() []string
}

type BarFeed interface {
	Feed
	GetCurrentBars() Bars
	GetLastBar() Bar
	GetNextBars() Bars
	GetCurrentDateTime() *time.Time
	BarsHaveAdjClose() bool
	GetFrequencies() []Frequency
	GetDefaultInstrument() string
	GetRegisteredInstruments() []string
	RegisterInstrument(instrument string, freq Frequency) error
	GetDataSeries(instrument string, freq Frequency) *series.Series
}
