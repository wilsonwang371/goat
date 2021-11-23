package strategy

import (
	"fmt"
	"goalgotrade/common"
	"goalgotrade/core"
	"sync"
	"time"

	lg "goalgotrade/logger"
)

type Analyzer interface {
	BeforeAttach(s Strategy) error
	Attached(s Strategy) error
	BeforeOnBars(s Strategy, bars common.Bars) error
}

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
	RegisterPositionOrder(position Position, order common.Order) error
	UnregisterPositionOrder(position Position, order common.Order) error
	UnregisterPosition(position Position) error
}

type Strategy interface {
	strategyLogger
	analyzeProvider
	positionCtrl
	OnStart() error
	OnIdle() error
	OnFinish(bars common.Bars) error
	OnOrderUpdated(order common.Order) error
	OnBars(bars common.Bars) error
	GetBarsProcessedEvent() common.Event
	GetFeed() common.Feed
	GetBroker() common.Broker
	SetBroker(broker common.Broker)
	GetUseAdjustedValues() bool
	GetLastPrice(instrument string) (float64, error)
	GetCurrentDateTime() *time.Time
	Run() (<-chan struct{}, error)
	Stop() error
}

type baseStrategy struct {
	Self interface{}

	mu         sync.RWMutex
	dispatcher common.Dispatcher
	broker     common.Broker
	barFeed    common.BarFeed

	barsProcessedEvent common.Event
	orderToPosition    map[uint64]Position
	activePositions    []Position
	namedAnalyzer      map[string]Analyzer
}

func NewBaseStrategy(bf common.BarFeed, bk common.Broker) *baseStrategy {
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
	return nil
}

func (s *baseStrategy) OnIdle() error {
	return nil
}

func (s *baseStrategy) OnOrderUpdated(order common.Order) error {
	return nil
}

func (s *baseStrategy) OnFinish(bars common.Bars) error {
	return nil
}

func (s *baseStrategy) OnBars(bars common.Bars) error {
	return nil
}

func (s *baseStrategy) RegisterPositionOrder(position Position, order common.Order) error {
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

func (s *baseStrategy) UnregisterPositionOrder(position Position, order common.Order) error {
	if _, ok := s.orderToPosition[order.GetId()]; ok {
		delete(s.orderToPosition, order.GetId())
	} else {
		return fmt.Errorf("invalid order to find")
	}
	return nil
}

func (s *baseStrategy) UnregisterPosition(position Position) error {
	if position.IsOpen() {
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
	orderEvent := args[1].(*common.OrderEvent)
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
	if len(args) != 1 {
		msg := "invalid amount of arguments"
		lg.Logger.Error(msg)
		panic(msg)
	}
	bars := args[0].(common.Bars)
	err := s.Self.(Strategy).OnBars(bars)
	if err != nil {
		return err
	}
	s.barsProcessedEvent.Emit(bars)
	return nil
}

func (s *baseStrategy) Run() (<-chan struct{}, error) {
	err := s.broker.GetOrderUpdatedEvent().Subscribe(func(args ...interface{}) error {
		return s.onOrderEvent(args)
	})
	if err != nil {
		return nil, err
	}

	err = s.barFeed.GetNewValuesEvent().Subscribe(func(args ...interface{}) error {
		return s.onBars(args)
	})
	if err != nil {
		return nil, err
	}

	err = s.dispatcher.GetStartEvent().Subscribe(func(args ...interface{}) error {
		return s.Self.(Strategy).OnStart()
	})
	if err != nil {
		return nil, err
	}

	err = s.dispatcher.GetIdleEvent().Subscribe(func(args ...interface{}) error {
		return s.Self.(Strategy).OnIdle()
	})
	if err != nil {
		return nil, err
	}

	err = s.dispatcher.AddSubject(s.broker)
	if err != nil {
		return nil, err
	}

	err = s.dispatcher.AddSubject(s.barFeed)
	if err != nil {
		return nil, err
	}

	ch, err := s.dispatcher.Run()
	if err != nil {
		return ch, err
	}

	currentBars := s.barFeed.GetCurrentBars()

	if currentBars != nil {
		if err := s.Self.(Strategy).OnFinish(currentBars); err != nil {
			return ch, err
		}
	} else {
		lg.Logger.Error("Feed was empty")
	}
	return ch, nil
}

func (s *baseStrategy) Stop() error {
	return s.dispatcher.Stop()
}

func (s *baseStrategy) GetBarsProcessedEvent() common.Event {
	return s.barsProcessedEvent
}

func (s *baseStrategy) GetFeed() common.Feed {
	return s.barFeed
}

func (s *baseStrategy) GetBroker() common.Broker {
	return s.broker
}

func (s *baseStrategy) SetBroker(broker common.Broker) {
	s.broker = broker
}

func (s *baseStrategy) GetUseAdjustedValues() bool {
	return false
}

func (s *baseStrategy) GetLastPrice(instrument string) (float64, error) {
	barList := s.Self.(Strategy).GetFeed().(common.BarFeed).GetLastBar(instrument)
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
		a.BeforeAttach(s)
		s.namedAnalyzer[name] = a
		a.Attached((s))
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
