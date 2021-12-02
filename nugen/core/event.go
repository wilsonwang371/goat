package core

import (
	"fmt"
	"reflect"
	"sync"
)

// EventType ...
type EventType int

// EventTypeBasic ...
const (
	EventTypeBasic EventType = iota
)

// Event ...
type Event interface {
	Name() string
	Type() EventType
	Data() map[string]interface{}
	Get(key string) (interface{}, bool)
	Set(key string, val interface{})
}

// EventHandler ...
type EventHandler func(event Event) error

// Channel ...
type Channel interface {
	Subscribe(handler EventHandler) error
	Unsubscribe(handler EventHandler) error
	Emit(event Event) []error
}

type channel struct {
	mu       sync.RWMutex
	handlers []EventHandler
}

// Subscribe ...
func (c *channel) Subscribe(handler EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, v := range c.handlers {
		if reflect.ValueOf(v).Pointer() == reflect.ValueOf(handler).Pointer() {
			return fmt.Errorf("duplicated")
		}
	}

	c.handlers = append(c.handlers, handler)
	return nil
}

// Unsubscribe ...
func (c *channel) Unsubscribe(handler EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < len(c.handlers); i++ {
		if reflect.ValueOf(c.handlers[i]).Pointer() == reflect.ValueOf(handler).Pointer() {
			c.handlers = append(c.handlers[:i], c.handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("not found")
}

// Emit ...
func (c *channel) Emit(event Event) []error {
	var res []error

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, v := range c.handlers {
		if err := v(event); err != nil {
			res = append(res, err)
		}
	}
	if len(res) > 0 {
		return res
	}
	return nil
}

// NewChannel ...
func NewChannel() Channel {
	return &channel{}
}

// BasicEvent ...
type BasicEvent struct {
	name string
	data map[string]interface{}
}

// Name ...
func (b *BasicEvent) Name() string {
	return b.name
}

// Type ...
func (b *BasicEvent) Type() EventType {
	return EventTypeBasic
}

// Data ...
func (b *BasicEvent) Data() map[string]interface{} {
	return b.data
}

// Get ...
func (b *BasicEvent) Get(key string) (interface{}, bool) {
	val, ok := b.data[key]
	return val, ok
}

// Set ...
func (b *BasicEvent) Set(key string, val interface{}) {
	b.data[key] = val
}

// NewBasicEvent ...
func NewBasicEvent(name string, data map[string]interface{}) Event {
	res := &BasicEvent{
		name: name,
		data: data,
	}
	if data == nil {
		res.data = make(map[string]interface{})
	}
	return res
}
