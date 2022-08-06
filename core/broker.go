package core

import "time"

type Broker interface {
	Subject
	GetOrderUpdatedEvent() Event
}

type dummyBroker struct {
	orderUpdatedEvent Event
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
func (e *dummyBroker) PeekDateTime() time.Time {
	return time.Now().UTC()
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

func NewDummyBroker() Broker {
	return &dummyBroker{
		orderUpdatedEvent: NewEvent(),
	}
}
