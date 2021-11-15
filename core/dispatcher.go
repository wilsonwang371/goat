package core

import (
	"sync"
	"time"

	"goalgotrade/common"
	lg "goalgotrade/logger"

	"go.uber.org/zap"
)

type dispatcher struct {
	mu              sync.RWMutex
	subjects        []common.Subject
	stopc           chan struct{}
	currentDateTime *time.Time

	startEvent common.Event
	idleEvent  common.Event
}

func NewDispatcher() common.Dispatcher {
	return &dispatcher{
		subjects:        []common.Subject{},
		stopc:           make(chan struct{}, 1),
		currentDateTime: nil,
		startEvent:      NewEvent(),
		idleEvent:       NewEvent(),
	}
}

func (d *dispatcher) AddSubject(subject common.Subject) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.subjects = append(d.subjects, subject)
	return nil
}

func (d *dispatcher) GetSubjects() []common.Subject {
	d.mu.RLock()
	defer d.mu.RUnlock()
	res := make([]common.Subject, len(d.subjects))
	for i, v := range d.subjects {
		res[i] = v
	}
	return res
}

func (d *dispatcher) GetStartEvent() common.Event {
	return d.startEvent
}

func (d *dispatcher) GetIdleEvent() common.Event {
	return d.idleEvent
}

func (d *dispatcher) GetCurrentDateTime() *time.Time {
	return d.currentDateTime
}

func (d *dispatcher) Stop() error {
	d.stopc <- struct{}{}
	return nil
}

func (d *dispatcher) cleanup() {
	for _, v := range d.subjects {
		v.Stop()
	}
	for _, v := range d.subjects {
		v.Join()
	}
}

func (d *dispatcher) dispatch() (eof bool, eventsDispatched bool) {
	eof = true
	eventsDispatched = false
	var smallestDateTime *time.Time = nil

	for _, v := range d.subjects {
		if !v.Eof() {
			eof = false
			t := v.PeekDateTime()
			if smallestDateTime == nil {
				smallestDateTime = t
			} else if smallestDateTime.After(*t) {
				smallestDateTime = t
			}
		}
	}

	if !eof {
		d.currentDateTime = smallestDateTime
		for _, v := range d.subjects {
			done, err := v.Dispatch()
			if err != nil {
				lg.Logger.Error("subject dispatch failed", zap.Error(err))
			}
			if done {
				eventsDispatched = true
			}
		}
	}
	return
}

func (d *dispatcher) Run() error {
	for _, v := range d.subjects {
		if err := v.Start(); err != nil {
			d.cleanup()
			return err
		}
	}

	d.startEvent.Emit()

	for {
		select {
		case <-d.stopc:
			d.currentDateTime = nil
			return nil
		default:
		}
		eof, eventDispatched := d.dispatch()
		if eof {
			d.stopc <- struct{}{}
		} else if eventDispatched {
			d.idleEvent.Emit()
		}
	}
}
