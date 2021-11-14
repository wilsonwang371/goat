package common

import "time"

type Dispatcher interface {
	AddSubject(subject Subject) error
	GetSubjects() []Subject

	GetStartEvent() Event
	GetIdleEvent() Event
	GetCurrentDateTime() *time.Time

	Stop() error
	Run() error
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
	PeekDateTime() time.Time

	GetDispatchPriority() int
	SetDispatchPriority(priority int)

	OnDispatcherRegistered(dispatcher Dispatcher) error
}

type EventHandler func(args ...interface{}) error

type Bar interface {
	SetUseAdjustedValue(useAdjusted bool) error
	GetUseAdjValue() bool

	GetDateTime() time.Time
	Open(adjusted bool) float64
	High(adjusted bool) float64
	Low(adjusted bool) float64
	Close(adjusted bool) float64
	Volume() int
	AdjClose() float64
	Frequency() float64
	Price() float64
}
