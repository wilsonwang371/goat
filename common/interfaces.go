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

// Bar, BarList, Bars are 3 different things
// Bar is an instance of Bar
// BarList is an array of Bars
// Bars is an instance of Bars which contains a map of BarLists
// However, due to compatibility issue with pyalgotrade, we use Bar and BarList interchangeably

type Bar interface {
	SetUseAdjustedValue(useAdjusted bool) error
	GetUseAdjValue() bool
	GetDateTime() *time.Time
	Open() float64
	High() float64
	Low() float64
	Close() float64
	Volume() float64
	AdjClose() float64
	Frequency() Frequency
	Price() float64
}

type Bars interface {
	GetDateTime() *time.Time
	GetInstruments() []string
	GetBarList(instrument string) []Bar
	GetFrequencies() []Frequency
	AddBarList(instrument string, barList []Bar) error
}

type Feed interface {
	Subject
	CreateDataSeries(key string, maxLen int) BarDataSeries
	GetNextValues() (*time.Time, Bars, []Frequency, error)
	GetNextValuesAndUpdateDS() (*time.Time, Bars, []Frequency, error)
	RegisterDataSeries(key string, freq Frequency) error
	GetNewValuesEvent() Event
	Reset()
	GetKeys() []string
	GetMaxLen() int
	IsLive() bool
}

type BarFeed interface {
	Feed
	GetCurrentBars() Bars
	GetLastBar(instrument string) []Bar
	GetNextBars() (Bars, error)
	GetCurrentDateTime() *time.Time
	BarsHaveAdjClose() bool
	GetFrequencies() []Frequency
	GetDefaultInstrument() string
	GetRegisteredInstruments() []string
	RegisterInstrument(instrument string, freq Frequency) error
	GetDataSeries(instrument string, freq Frequency) *series.Series
}

type BarDataSeries interface {
	Append(bar Bar) error
	AppendWithDateTime(dateTime time.Time, bar Bar) error
	OpenDS() *series.Series
	HighDS() *series.Series
	LowDS() *series.Series
	CloseDS() *series.Series
	AdjCloseDS() *series.Series
	VolumeDS() *series.Series
	PriceDS() *series.Series
	ExtraDS() map[string]series.Series
}
