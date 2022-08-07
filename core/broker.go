package core

import (
	"fmt"
	"goalgotrade/logger"
	"time"

	"go.uber.org/zap"
)

type Broker interface {
	Subject
	GetOrderUpdatedEvent() Event
}

type dummyBroker struct {
	orderUpdatedEvent Event
	datafeed          DataFeed
}

// Dispatch implements Broker
func (e *dummyBroker) Dispatch() bool {
	return true
}

// Eof implements Broker
func (e *dummyBroker) Eof() bool {
	return true
}

// Join implements Broker
func (e *dummyBroker) Join() error {
	return nil
}

// PeekDateTime implements Broker
func (e *dummyBroker) PeekDateTime() *time.Time {
	t := time.Now().UTC()
	return &t
}

// Start implements Broker
func (e *dummyBroker) Start() error {
	return nil
}

// Stop implements Broker
func (e *dummyBroker) Stop() error {
	return nil
}

// GetOrderUpdatedEvent implements Broker
func (e *dummyBroker) GetOrderUpdatedEvent() Event {
	return e.orderUpdatedEvent
}

func (e *dummyBroker) onBars(args ...interface{}) error {
	logger.Logger.Info("broker onBars")
	if len(args) != 2 {
		return fmt.Errorf("onBars args length should be 2")
	}

	currentTime := args[0].(time.Time)
	data := args[1].(map[string]interface{})
	bars := make(map[string]Bar, len(data))
	for k, v := range data {
		bars[k] = v.(Bar)
	}

	logger.Logger.Info("onBars",
		zap.Time("time", currentTime),
		zap.Any("bars", bars))

	// TODO: implement fill strategy

	return nil
}

func NewDummyBroker(feed DataFeed) Broker {
	broker := &dummyBroker{
		orderUpdatedEvent: NewEvent(),
		datafeed:          feed,
	}
	feed.GetNewValueEvent().Subscribe(broker.onBars)
	return broker
}
