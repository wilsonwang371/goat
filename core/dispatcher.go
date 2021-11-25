package core

import (
	"goalgotrade/common"
	"sync"
	"time"

	lg "goalgotrade/logger"

	"go.uber.org/zap"
)

type dispatcher struct {
	mu              sync.RWMutex
	subjects        []common.Subject
	stopC           chan struct{}
	currentDateTime *time.Time

	startEvent common.Event
	idleEvent  common.Event
}

func NewDispatcher() common.Dispatcher {
	return &dispatcher{
		subjects:        []common.Subject{},
		stopC:           make(chan struct{}, 1),
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
	d.stopC <- struct{}{}
	return nil
}

func (d *dispatcher) cleanup() {
	for _, v := range d.subjects {
		if err := v.Stop(); err != nil {
			lg.Logger.Warn("error", zap.Error(err))
		}
	}
	for _, v := range d.subjects {
		if err := v.Join(); err != nil {
			lg.Logger.Warn("error", zap.Error(err))
		}
	}
}

func (d *dispatcher) dispatch() (eof bool, eventsDispatched bool) {
	eof = true
	eventsDispatched = false
	var smallestDateTime *time.Time = nil
	wg := sync.WaitGroup{}

	for _, v := range d.subjects {
		if !v.Eof() {
			eof = false
			t := v.PeekDateTime()
			if t == nil {
				continue
			}
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
			wg.Add(1)
			go func(sub common.Subject) {
				defer wg.Done()
				done, err := sub.Dispatch()
				if err != nil {
					lg.Logger.Error("subject dispatch failed", zap.Error(err))
				}
				if done {
					eventsDispatched = true
				}
			}(v)
		}
		wg.Wait()
	}
	return
}

func (d *dispatcher) Run() (<-chan struct{}, error) {
	ch := make(chan struct{}, 1)
	go func() {
		d.mainDispatchLoop()
		ch <- struct{}{}
	}()
	return ch, nil
}

func (d *dispatcher) mainDispatchLoop() {
	d.mu.RLock()
	for _, v := range d.subjects {
		if err := v.Start(); err != nil {
			d.cleanup()
			d.mu.RUnlock()
			lg.Logger.Error("error starting subjects", zap.Error(err))
			panic(err)
		}
	}
	d.mu.RUnlock()

	d.startEvent.Emit()

	for {
		select {
		case <-d.stopC:
			d.currentDateTime = nil
			return
		default:
		}
		d.mu.RLock()
		eof, eventDispatched := d.dispatch()
		d.mu.RUnlock()
		if eof {
			d.stopC <- struct{}{}
		} else if !eventDispatched {
			d.idleEvent.Emit()
		}
	}
}
