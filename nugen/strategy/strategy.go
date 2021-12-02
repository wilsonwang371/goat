package strategy

import (
	"fmt"
	"goalgotrade/common"
	"goalgotrade/core"
	"goalgotrade/nugen/bar"
	"goalgotrade/nugen/broker"
	"goalgotrade/nugen/feed"
	"goalgotrade/nugen/feed/barfeed"
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
	GetNamedAnalyzer(name string) (Analyzer, error)
}

type positionCtrl interface {
	RegisterPositionOrder(position Position, order *broker.Order) error
	UnregisterPositionOrder(position Position, order *broker.Order) error
	UnregisterPosition(position Position) error
}

type strategyEvent interface {
	OnStart() error
	OnIdle() error
	OnFinish(bars *bar.Bars) error
	OnOrderUpdated(order *broker.Order) error
	OnBars(bars *bar.Bars) error
}

type positionNotification interface {
	OnEnterOk(p Position) error
	OnEnterCanceled(p Position) error
	OnExitOk(p Position) error
	OnExitCanceled(p Position) error
}

type Strategy interface {
	strategyLogger
	analyzeProvider
	positionCtrl
	strategyEvent
	positionNotification
	GetBarsProcessedEvent() common.Event
	GetFeed() feed.Feed
	GetBroker() *broker.Broker
	SetBroker(broker *broker.Broker)
	GetUseAdjustedValues() bool
	GetLastPrice(instrument string) (float64, error)
	GetCurrentDateTime() *time.Time
	Run() error
	Stop() error
}

type baseStrategy struct {
	Self       interface{}
	mu         sync.RWMutex
	dispatcher common.Dispatcher
	broker     *broker.Broker
	barFeed    barfeed.BarFeed

	barsProcessedEvent common.Event
	orderToPosition    map[uint64]Position
	activePositions    []Position
	namedAnalyzer      map[string]Analyzer
}

func NewBaseStrategy(bf barfeed.BarFeed, bk *broker.Broker) *baseStrategy {
	res := &baseStrategy{
		dispatcher:         core.NewDispatcher(),
		barFeed:            bf,
		broker:             bk,
		barsProcessedEvent: core.NewEvent(),
		activePositions:    []Position{},
		namedAnalyzer:      map[string]Analyzer{},
	}
	res.Self = res
	return res
}

func (s *baseStrategy) OnStart() error {
	lg.Logger.Debug("OnStart()")
	return nil
}

func (s *baseStrategy) OnIdle() error {
	lg.Logger.Debug("OnIdle()")
	return nil
}

func (s *baseStrategy) OnOrderUpdated(order *broker.Order) error {
	lg.Logger.Debug("OnOrderUpdated()")
	return nil
}

func (s *baseStrategy) OnFinish(bars *bar.Bars) error {
	lg.Logger.Debug("OnFinish()")
	return nil
}

func (s *baseStrategy) OnBars(bars *bar.Bars) error {
	fmt.Fprintf(os.Stderr, "OnBars %s bars %v\n", bars.GetInstruments(), bars)
	return nil
}

func (s *baseStrategy) RegisterPositionOrder(position Position, order *broker.Order) error {
	if !order.IsActive() {
		return fmt.Errorf("registering an inactive order")
	}
	for _, v := range s.activePositions {
		if v == position {
			return fmt.Errorf("position exists already")
		}
	}
	s.activePositions = append(s.activePositions, position)

	if _, ok := s.orderToPosition[order.GetId()]; !ok {
		s.orderToPosition[order.GetId()] = position
	} else {
		return fmt.Errorf("order exists already")
	}
	return nil
}

func (s *baseStrategy) UnregisterPositionOrder(position Position, order *broker.Order) error {
	if _, ok := s.orderToPosition[order.GetId()]; ok {
		delete(s.orderToPosition, order.GetId())
	} else {
		return fmt.Errorf("invalid order to find")
	}
	return nil
}

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

func (s *baseStrategy) onOrderEvent(args ...interface{}) error {
	if len(args) != 2 {
		msg := "invalid number of arguments"
		lg.Logger.Error(msg)
		panic(msg)
	}
	// bk := args[0].(broker.Broker)
	orderEvent := args[1].(*broker.OrderEvent)
	order := orderEvent.Order
	err := s.Self.(Strategy).OnOrderUpdated(order)
	if err != nil {
		return err
	}

	if pos, ok := s.orderToPosition[order.GetId()]; ok {
		if pos == nil {
			msg := "invalid position"
			lg.Logger.Error(msg)
			panic(msg)
		}
		if order.IsActive() {
			err := s.Self.(Strategy).UnregisterPositionOrder(pos, order)
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

func (s *baseStrategy) onBars(args ...interface{}) error {
	if len(args) != 2 {
		msg := "invalid amount of arguments"
		lg.Logger.Error(msg)
		panic(msg)
	}
	dateTime := args[0].(*time.Time) // not used
	if dateTime == nil {
		lg.Logger.Error("invalid date time")
	}
	bars := args[1].(common.Bars)
	err := s.Self.(Strategy).OnBars(bars)
	if err != nil {
		return err
	}
	s.barsProcessedEvent.Emit(bars)
	return nil
}

func (s *baseStrategy) Run() error {
	err := s.broker.GetOrderUpdatedEvent().Subscribe(func(args ...interface{}) error {
		return s.onOrderEvent(args)
	})
	if err != nil {
		return err
	}

	err = s.barFeed.GetNewValuesEvent().Subscribe(func(args ...interface{}) error {
		return s.onBars(args...)
	})
	if err != nil {
		return err
	}

	err = s.dispatcher.GetStartEvent().Subscribe(func(args ...interface{}) error {
		return s.Self.(Strategy).OnStart()
	})
	if err != nil {
		return err
	}

	err = s.dispatcher.GetIdleEvent().Subscribe(func(args ...interface{}) error {
		return s.Self.(Strategy).OnIdle()
	})
	if err != nil {
		return err
	}

	err = s.dispatcher.AddSubject(s.broker)
	if err != nil {
		return err
	}

	err = s.dispatcher.AddSubject(s.barFeed)
	if err != nil {
		return err
	}

	ch, err := s.dispatcher.Run()
	if err != nil {
		return err
	}

	<-ch

	currentBars := s.barFeed.GetCurrentBars()

	if currentBars != nil {
		if err := s.Self.(Strategy).OnFinish(currentBars); err != nil {
			return err
		}
	} else {
		lg.Logger.Info("Feed was empty")
	}
	return nil
}

func (s *baseStrategy) Stop() error {
	lg.Logger.Info("stopping strategy")
	return s.dispatcher.Stop()
}

func (s *baseStrategy) GetBarsProcessedEvent() common.Event {
	return s.barsProcessedEvent
}

func (s *baseStrategy) GetFeed() feed.Feed {
	return s.barFeed
}

func (s *baseStrategy) GetBroker() *broker.Broker {
	return s.broker
}

func (s *baseStrategy) SetBroker(broker *broker.Broker) {
	s.broker = broker
}

func (s *baseStrategy) GetUseAdjustedValues() bool {
	return false
}

func (s *baseStrategy) GetLastPrice(instrument string) (float64, error) {
	barList := s.Self.(Strategy).GetFeed().(barfeed.BarFeed).GetLastBar(instrument)
	if barList == nil {
		return 0, fmt.Errorf("invalid bar after calling GetLastBar")
	}
	if len(barList) != 1 {
		return 0, fmt.Errorf("too many bars getting from GetLastBar")
	}
	return barList[0].Price(), nil
}

func (s *baseStrategy) GetCurrentDateTime() *time.Time {
	return s.barFeed.GetCurrentDateTime()
}

func (s *baseStrategy) Debug(msg string) {
	lg.Logger.Debug(msg)
}

func (s *baseStrategy) Info(msg string) {
	lg.Logger.Info(msg)
}

func (s *baseStrategy) Error(msg string) {
	lg.Logger.Error(msg)
}

func (s *baseStrategy) Warning(msg string) {
	lg.Logger.Warn(msg)
}

func (s *baseStrategy) Critical(msg string) {
	lg.Logger.Fatal(msg)
}

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

func (s *baseStrategy) GetNamedAnalyzer(name string) (Analyzer, error) {
	if a, ok := s.namedAnalyzer[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("analyzer not found")
}

func (s *baseStrategy) OnEnterOk(p Position) error {
	panic("not implemented")
	return nil
}

func (s *baseStrategy) OnEnterCanceled(p Position) error {
	panic("not implemented")
	return nil
}

func (s *baseStrategy) OnExitOk(p Position) error {
	panic("not implemented")
	return nil
}

func (s *baseStrategy) OnExitCanceled(p Position) error {
	panic("not implemented")
	return nil
}
