package core

import (
	lg "goalgotrade/nugen/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Subject ...
type Subject interface {
	Start() error
	Stop() error
	Join() error
	Eof() bool
	Dispatch(subject interface{}) (bool, error)
	PeekDateTime() *time.Time
	GetDispatchPriority() int
	SetDispatchPriority(priority int)
	OnDispatcherRegistered(dispatcher Dispatcher) error
}

// Dispatcher ...
type Dispatcher interface {
	AddSubject(subject Subject)
	Run()
	Stop() error
	StartChannel() Channel
	IdleChannel() Channel
	CurrentTime() *time.Time
}

type dispatcher struct {
	mu       sync.RWMutex
	subjects []Subject
	stopC    chan struct{}

	currentTime  *time.Time
	startChannel Channel
	idleChannel  Channel
}

// NewDispatcher ...
func NewDispatcher() Dispatcher {
	return &dispatcher{
		subjects:     []Subject{},
		stopC:        make(chan struct{}, 1),
		currentTime:  nil,
		startChannel: NewChannel(),
		idleChannel:  NewChannel(),
	}
}

// AddSubject ...
func (d *dispatcher) AddSubject(subject Subject) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.subjects = append(d.subjects, subject)
}

// Stop ...
func (d *dispatcher) Stop() error {
	d.stopC <- struct{}{}
	return nil
}

// Run ...
func (d *dispatcher) Run() {
	d.mainDispatchLoop()
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
	var smallestTime *time.Time = nil
	wg := sync.WaitGroup{}

	for _, v := range d.subjects {
		if !v.Eof() {
			eof = false
			t := v.PeekDateTime()
			if t == nil {
				continue
			}
			if smallestTime == nil {
				smallestTime = t
			} else if smallestTime.After(*t) {
				smallestTime = t
			}
		}
	}

	if !eof {
		d.currentTime = smallestTime
		for _, v := range d.subjects {
			wg.Add(1)
			go func(sub Subject) {
				defer wg.Done()
				done, err := sub.Dispatch(sub)
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

	d.StartChannel().Emit(NewBasicEvent("start", nil))

	for {
		select {
		case <-d.stopC:
			d.currentTime = nil
			return
		default:
		}
		d.mu.RLock()
		eof, eventDispatched := d.dispatch()
		d.mu.RUnlock()
		if eof {
			d.stopC <- struct{}{}
		} else if !eventDispatched {
			d.IdleChannel().Emit(NewBasicEvent("idle", nil))
		}
	}
}

// StartChannel ...
func (d *dispatcher) StartChannel() Channel {
	return d.startChannel
}

// IdleChannel ...
func (d *dispatcher) IdleChannel() Channel {
	return d.idleChannel
}

// CurrentTime ...
func (d *dispatcher) CurrentTime() *time.Time {
	return d.currentTime
}
