package core

import (
	"fmt"
	"goalgotrade/logger"
	"time"

	"go.uber.org/zap"
)

type Bar interface{}

type Bars interface{}

type Order interface{}

type OrderEvent interface {
	GetOrder() Order
	GetEventType()
	GetEventInfo()
}

type StrategyEventListener interface {
	OnStart(args ...interface{}) error
	OnIdle() error
	OnFinish(args ...interface{}) error
	OnBars(bars Bars) error
	OnOrderUpdated(order Order) error
	OnOrderEvent(orderEvent OrderEvent) error
}

type StrategyController interface{}

type strategyController struct {
	listener StrategyEventListener
	broker   Broker
	dataFeed DataFeed

	dispatcher Dispatcher

	barProcessedEvent Event
}

func (s *strategyController) onStart(args ...interface{}) error {
	logger.Logger.Info("onStart")
	return s.listener.OnStart()
}

func (s *strategyController) onIdle(args ...interface{}) error {
	logger.Logger.Info("onIdle")
	/*
			# Force a resample check to avoid depending solely on the underlying
		        # barfeed events.
		        for resampledBarFeed in self.__resampledBarFeeds:
		            resampledBarFeed.checkNow(self.getCurrentDateTime())
	*/
	return s.listener.OnIdle()
}

func (s *strategyController) onBars(args ...interface{}) error {
	logger.Logger.Info("onBars")
	if len(args) != 2 {
		return fmt.Errorf("onBars args length should be 2")
	}

	currentTime := args[0].(time.Time)
	bars := args[1].(Bars)

	logger.Logger.Info("onBars",
		zap.Time("time", currentTime),
		zap.Any("bars", bars))

	s.listener.OnBars(bars)
	s.barProcessedEvent.Emit(bars)

	return nil
}

func (s *strategyController) onOrderEvent(args ...interface{}) error {
	logger.Logger.Info("onOrderEvent")
	if len(args) != 2 {
		return fmt.Errorf("onOrderEvent args length should be 2")
	}

	broker := args[0].(Broker)
	orderEvent := args[1].(OrderEvent)

	logger.Logger.Info("onOrderEvent",
		zap.Any("broker", broker),
		zap.Any("orderEvent", orderEvent))

	s.listener.OnOrderUpdated(orderEvent.GetOrder())

	//TODO: handle order event
	/*
			 # Notify the position about the order event.
		        pos = self.__orderToPosition.get(order.getId(), None)
		        if pos is not None:
		            # Unlink the order from the position if its not active anymore.
		            if not order.isActive():
		                self.unregisterPositionOrder(pos, order)

		            pos.onOrderEvent(orderEvent)
	*/

	s.listener.OnOrderEvent(orderEvent)

	return nil
}

func NewStrategyController(strategyEventListener StrategyEventListener,
	broker Broker, dataFeed DataFeed) StrategyController {
	controller := &strategyController{
		listener:          strategyEventListener,
		broker:            broker,
		dataFeed:          dataFeed,
		dispatcher:        NewDispatcher(),
		barProcessedEvent: NewEvent(),
	}

	controller.dispatcher.AddSubject(controller.broker)
	controller.dispatcher.AddSubject(controller.dataFeed)

	controller.dispatcher.GetStartEvent().Subscribe(controller.onStart)
	controller.dispatcher.GetIdleEvent().Subscribe(controller.onIdle)

	controller.dataFeed.GetNewValueEvent().Subscribe(controller.onBars)
	controller.broker.GetOrderUpdatedEvent().Subscribe(controller.onOrderEvent)

	return controller
}
