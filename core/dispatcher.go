package core

import (
	"fmt"
	"reflect"
	"time"
)

// Subject interface
type Subject interface {
	Start() error
	Stop() error
	Join() error
	Eof() bool
	Dispatch() bool
	PeekDateTime() time.Time
}

// Dispatcher interface
type Dispatcher interface {
	AddSubject(subject Subject)
	Run()
	Stop()
	GetSubjects() []Subject
}

// EventHandler interface
type EventHandler interface {
	handle(args ...interface{}) error
}

// Event interface
type Event interface {
	Subscribe(handler EventHandler) error
	Unsubscribe(handler EventHandler) error
	Emit(args ...interface{})
}

type dispatcher struct {
	subjects   []Subject
	stopC      chan struct{}
	startEvent Event
	idleEvent  Event
	lastTime   time.Time
}

func NewDispatcher() Dispatcher {
	return &dispatcher{
		stopC:      make(chan struct{}),
		startEvent: NewEvent(),
		idleEvent:  NewEvent(),
	}
}

func (d *dispatcher) AddSubject(subject Subject) {
	d.subjects = append(d.subjects, subject)
}

func (d *dispatcher) dispatchSubject(subject Subject, smallestTime time.Time) bool {
	if !subject.Eof() && subject.PeekDateTime().Before(smallestTime) {
		return subject.Dispatch()
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
			if smallestNewTime == nil || newTime.Before(*smallestNewTime) {
				smallestNewTime = &newTime
			}
		}
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
			return
		default:
			if eof, dispatched := d.dispatch(); eof {
				d.stopC <- struct{}{}
			} else {
				if !dispatched {
					d.idleEvent.Emit()
				}
			}
		}
	}
}

func (d *dispatcher) Stop() {
	close(d.stopC)
}

func (d *dispatcher) GetSubjects() []Subject {
	return d.subjects
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
		handler.handle(args...)
	}
}
