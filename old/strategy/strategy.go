package strategy

import (
	"fmt"
	"goalgotrade/bar"
	"goalgotrade/broker"
	"goalgotrade/core"
	"goalgotrade/feed"
	"os"
	"sync"
	"time"

	lg "goalgotrade/logger"
)

type strategyLogger interface {
	Critical(msg string)
	Warning(msg string)
	Error(msg string)
	Info(msg string)
	Debug(msg string)
}

type analyzeProvider interface {
	AttachAnalyzer(a Analyzer, name string) error
	NamedAnalyzer(name string) (Analyzer, error)
}

type positionCtrl interface {
	RegisterPositionOrder(position Position, order broker.Order) error
	UnregisterPositionOrder(position Position, order broker.Order) error
	UnregisterPosition(position Position) error
}

type strategyEvent interface {
	OnStart() error
	OnIdle() error
	OnFinish(bars bar.Bars) error
	OnOrderUpdated(order broker.Order) error
	OnBars(bars bar.Bars) error
}

type positionNotification interface {
	OnEnterOk(p Position) error
	OnEnterCanceled(p Position) error
	OnExitOk(p Position) error
	OnExitCanceled(p Position) error
}

// Strategy ...
type Strategy interface {
	strategyLogger
	analyzeProvider
	positionCtrl
	strategyEvent
	positionNotification
	BarsProcessedChannel() core.Channel
	Feed() feed.BaseFeed
	Broker() broker.Broker
	SetBroker(broker broker.Broker)
	UseAdjustedValues() bool
	LastPrice(b Strategy, instrument string) (float64, error)
	CurrentTime() *time.Time
	Run(b Strategy) error
	Stop() error
}

type baseStrategy struct {
	mu         sync.RWMutex
	dispatcher core.Dispatcher
	broker     broker.Broker
	barFeed    feed.BaseBarFeed

	barsProcessedChannel core.Channel
	orderToPosition      map[uint64]Position
	activePositions      []Position
	namedAnalyzer        map[string]Analyzer
}

// NewBaseStrategy ...
func NewBaseStrategy(bf feed.BaseBarFeed, bk broker.Broker) Strategy {
	return newBaseStrategy(bf, bk)
}

func newBaseStrategy(bf feed.BaseBarFeed, bk broker.Broker) *baseStrategy {
	res := &baseStrategy{
		dispatcher:           core.NewDispatcher(),
		barFeed:              bf,
		broker:               bk,
		barsProcessedChannel: core.NewChannel(),
		activePositions:      []Position{},
		namedAnalyzer:        map[string]Analyzer{},
	}
	return res
}

// OnStart ...
func (s *baseStrategy) OnStart() error {
	lg.Logger.Debug("baseStrategy Start() called")
	return nil
}

// OnIdle ...
func (s *baseStrategy) OnIdle() error {
	// lg.Logger.Debug("OnIdle()")
	return nil
}

// OnOrderUpdated ...
func (s *baseStrategy) OnOrderUpdated(order broker.Order) error {
	lg.Logger.Debug("OnOrderUpdated()")
	return nil
}

// OnFinish ...
func (s *baseStrategy) OnFinish(bars bar.Bars) error {
	lg.Logger.Debug("OnFinish()")
	return nil
}

// OnBars ...
func (s *baseStrategy) OnBars(bars bar.Bars) error {
	fmt.Fprintf(os.Stderr, "OnBars %s bars %v\n", bars.Instruments(), bars)
	return nil
}

// RegisterPositionOrder ...
func (s *baseStrategy) RegisterPositionOrder(position Position, order broker.Order) error {
	if !order.IsActive() {
		return fmt.Errorf("registering an inactive order")
	}
	for _, v := range s.activePositions {
		if v == position {
			return fmt.Errorf("position exists already")
		}
	}
	s.activePositions = append(s.activePositions, position)

	if _, ok := s.orderToPosition[order.Id()]; !ok {
		s.orderToPosition[order.Id()] = position
	} else {
		return fmt.Errorf("order exists already")
	}
	return nil
}

// UnregisterPositionOrder ...
func (s *baseStrategy) UnregisterPositionOrder(position Position, order broker.Order) error {
	if _, ok := s.orderToPosition[order.Id()]; ok {
		delete(s.orderToPosition, order.Id())
	} else {
		return fmt.Errorf("invalid order to find")
	}
	return nil
}

// UnregisterPosition ...
func (s *baseStrategy) UnregisterPosition(position Position) error {
	if position.IsOpen(position) {
		return fmt.Errorf("position is still open")
	}
	idx := -1
	for i, v := range s.activePositions {
		if v == position {
			idx = i
		}
	}
	if idx == -1 {
		return fmt.Errorf("position not found")
	}
	s.activePositions = append(s.activePositions[:idx], s.activePositions[idx+1:]...)
	return nil
}

func (s *baseStrategy) onOrderEvent(b Strategy, orderEvent *broker.OrderEvent) error {
	order := orderEvent.Order
	err := b.OnOrderUpdated(order)
	if err != nil {
		return err
	}

	if pos, ok := s.orderToPosition[order.Id()]; ok {
		if pos == nil {
			msg := "invalid position"
			lg.Logger.Error(msg)
			panic(msg)
		}
		if order.IsActive() {
			err := b.UnregisterPositionOrder(pos, order)
			if err != nil {
				return err
			}
		}
		err := pos.OnOrderEvent(orderEvent)
		if err != nil {
			return err
		}
	}
	return nil
}

// Run ...
func (s *baseStrategy) Run(b Strategy) error {
	err := s.broker.OrderUpdatedChannel().Subscribe(func(event core.Event) error {
		orderEventRaw, ok := event.Get("event")
		if !ok {
			panic("invalid event")
		}
		return s.onOrderEvent(b, orderEventRaw.(*broker.OrderEvent))
	})
	if err != nil {
		return err
	}

	err = s.barFeed.NewValueChannel().Subscribe(func(event core.Event) error {
		timeRaw, ok := event.Get("time")
		if !ok {
			panic("invalid time in event")
		}

		barsRaw, ok := event.Get("bars")
		if !ok {
			panic("invalid bars in event")
		}

		dateTime := timeRaw.(*time.Time)
		bars := barsRaw.(bar.Bars)

		if dateTime == nil {
			lg.Logger.Error("invalid date time")
		}
		err := b.OnBars(bars)
		if err != nil {
			return err
		}
		s.barsProcessedChannel.Emit(core.NewBasicEvent("bars-processed", map[string]interface{}{
			"value": bars,
		}))
		return nil
	})
	if err != nil {
		return err
	}

	err = s.dispatcher.StartChannel().Subscribe(func(event core.Event) error {
		lg.Logger.Debug("Start Event!")
		return b.OnStart()
	})
	if err != nil {
		return err
	}

	err = s.dispatcher.IdleChannel().Subscribe(func(event core.Event) error {
		// lg.Logger.Debug("Idle Event!")
		return b.OnIdle()
	})
	if err != nil {
		return err
	}

	s.dispatcher.AddSubject(s.broker)
	s.dispatcher.AddSubject(s.barFeed)

	s.dispatcher.Run()

	currentBars := s.barFeed.CurrentBars()

	if currentBars != nil {
		if err := b.OnFinish(currentBars); err != nil {
			return err
		}
	} else {
		lg.Logger.Info("Feed was empty")
	}
	return nil
}

// Stop ...
func (s *baseStrategy) Stop() error {
	lg.Logger.Info("stopping strategy")
	return s.dispatcher.Stop()
}

// BarsProcessedChannel ...
func (s *baseStrategy) BarsProcessedChannel() core.Channel {
	return s.barsProcessedChannel
}

// Feed ...
func (s *baseStrategy) Feed() feed.BaseFeed {
	return s.barFeed
}

// Broker ...
func (s *baseStrategy) Broker() broker.Broker {
	return s.broker
}

// SetBroker ...
func (s *baseStrategy) SetBroker(broker broker.Broker) {
	s.broker = broker
}

// UseAdjustedValues ...
func (s *baseStrategy) UseAdjustedValues() bool {
	return false
}

// LastPrice ...
func (s *baseStrategy) LastPrice(b Strategy, instrument string) (float64, error) {
	bar := b.Feed().(feed.BaseBarFeed).LastBar(instrument)
	if bar == nil {
		return 0, fmt.Errorf("invalid bar after calling GetLastBar")
	}
	return bar.Price(), nil
}

// CurrentTime ...
func (s *baseStrategy) CurrentTime() *time.Time {
	return s.barFeed.CurrentTime()
}

// Debug ...
func (s *baseStrategy) Debug(msg string) {
	lg.Logger.Debug(msg)
}

// Info ...
func (s *baseStrategy) Info(msg string) {
	lg.Logger.Info(msg)
}

// Error ...
func (s *baseStrategy) Error(msg string) {
	lg.Logger.Error(msg)
}

// Warning ...
func (s *baseStrategy) Warning(msg string) {
	lg.Logger.Warn(msg)
}

// Critical ...
func (s *baseStrategy) Critical(msg string) {
	lg.Logger.Fatal(msg)
}

// AttachAnalyzer ...
func (s *baseStrategy) AttachAnalyzer(a Analyzer, name string) error {
	if a == nil {
		return fmt.Errorf("analyzer is nil")
	}
	if _, ok := s.namedAnalyzer[name]; !ok {
		if err := a.BeforeAttach(s); err != nil {
			return fmt.Errorf("before attach analyzer failed: %v", err)
		}
		s.namedAnalyzer[name] = a
		if err := a.Attached(s); err != nil {
			return fmt.Errorf("attached analyzer failed: %v", err)
		}
		return nil
	}
	return fmt.Errorf("analyzer %s already exists", name)
}

// NamedAnalyzer ...
func (s *baseStrategy) NamedAnalyzer(name string) (Analyzer, error) {
	if a, ok := s.namedAnalyzer[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("analyzer not found")
}

// OnEnterOk ...
func (s *baseStrategy) OnEnterOk(p Position) error {
	panic("not implemented")
	return nil
}

// OnEnterCanceled ...
func (s *baseStrategy) OnEnterCanceled(p Position) error {
	panic("not implemented")
	return nil
}

// OnExitOk ...
func (s *baseStrategy) OnExitOk(p Position) error {
	panic("not implemented")
	return nil
}

// OnExitCanceled ...
func (s *baseStrategy) OnExitCanceled(p Position) error {
	panic("not implemented")
	return nil
}
