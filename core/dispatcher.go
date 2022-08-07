package core

import (
	"fmt"
	"goalgotrade/logger"
	"reflect"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Subject interface
type Subject interface {
	Start() error
	Stop() error
	Join() error
	Eof() bool
	Dispatch() bool
	PeekDateTime() *time.Time // it can be nil if no data is available
}

// Dispatcher interface
type Dispatcher interface {
	AddSubject(subject Subject)
	Run()
	Stop()
	GetSubjects() []Subject
	GetStartEvent() Event
	GetIdleEvent() Event
}

// EventHandler interface
type EventHandler func(args ...interface{}) error

// Event interface
type Event interface {
	Subscribe(handler EventHandler) error
	Unsubscribe(handler EventHandler) error
	Emit(args ...interface{})
}

type dispatcher struct {
	subjects   []Subject
	stopC      chan struct{}
	stopCMutex sync.Mutex
	isStopped  bool
	startEvent Event
	idleEvent  Event
	lastTime   time.Time
}

// GetIdleEvent implements Dispatcher
func (d *dispatcher) GetIdleEvent() Event {
	return d.idleEvent
}

// GetStartEvent implements Dispatcher
func (d *dispatcher) GetStartEvent() Event {
	return d.startEvent
}

func (d *dispatcher) AddSubject(subject Subject) {
	d.subjects = append(d.subjects, subject)
}

func (d *dispatcher) dispatchSubject(subject Subject, smallestTime time.Time) bool {
	t := subject.PeekDateTime()
	if t == nil {
		logger.Logger.Info("no data available yet", zap.String("subject", reflect.TypeOf(subject).String()))
		return false
	}
	if !subject.Eof() && !t.Before(smallestTime) {
		return subject.Dispatch()
	} else {
		logger.Logger.Info("data not dispatched", zap.Any("eof", subject.Eof()),
			zap.Any("smallestTime", smallestTime),
			zap.Any("PeekDateTime", t))
	}
	return false
}

func (d *dispatcher) dispatch() (eof bool, dispatched bool) {
	eof = true
	dispatched = false
	var smallestNewTime *time.Time

	for _, subject := range d.subjects {
		if !subject.Eof() {
			eof = false
			newTime := subject.PeekDateTime()
			if newTime != nil && (smallestNewTime == nil || newTime.Before(*smallestNewTime)) {
				smallestNewTime = newTime
			}
		}
	}

	if smallestNewTime == nil {
		// we dont have any data yet
		return
	}

	if !eof {
		if smallestNewTime != nil {
			d.lastTime = *smallestNewTime
		}
		for _, subject := range d.subjects {
			if d.dispatchSubject(subject, *smallestNewTime) {
				dispatched = true
			}
		}
	}
	return
}

func (d *dispatcher) Run() {
	for _, subject := range d.subjects {
		subject.Start()
	}
	d.startEvent.Emit()

	for {
		select {
		case <-d.stopC:
			for _, subject := range d.subjects {
				subject.Stop()
			}
			for _, subject := range d.subjects {
				subject.Join()
			}
			close(d.stopC)
			return
		default:
			if eof, dispatched := d.dispatch(); eof {
				d.Stop()
			} else {
				if !dispatched {
					d.idleEvent.Emit()
				}
			}
		}
	}
}

func (d *dispatcher) Stop() {
	d.stopCMutex.Lock()
	defer d.stopCMutex.Unlock()
	if !d.isStopped {
		return
	}
	d.stopC <- struct{}{}
	d.isStopped = true
}

func (d *dispatcher) GetSubjects() []Subject {
	return d.subjects
}

func NewDispatcher() Dispatcher {
	return &dispatcher{
		stopC:      make(chan struct{}, 2),
		stopCMutex: sync.Mutex{},
		startEvent: NewEvent(),
		idleEvent:  NewEvent(),
	}
}

type event struct {
	handlers []EventHandler
}

func NewEvent() Event {
	return &event{}
}

func (e *event) Subscribe(handler EventHandler) error {
	e.handlers = append(e.handlers, handler)
	return nil
}

func (e *event) Unsubscribe(handler EventHandler) error {
	for i := 0; i < len(e.handlers); i++ {
		if reflect.ValueOf(e.handlers[i]).Pointer() == reflect.ValueOf(handler).Pointer() {
			e.handlers = append(e.handlers[:i], e.handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("not found")
}

func (e *event) Emit(args ...interface{}) {
	for _, handler := range e.handlers {
		handler(args...)
	}
}
