package core

import (
	"fmt"
	"goalgotrade/common"
	"reflect"
	"sync"
)

type event struct {
	mu       sync.RWMutex
	handlers []common.EventHandler
}

func NewEvent() common.Event {
	return &event{
		handlers: []common.EventHandler{},
	}
}

func (e *event) Subscribe(handler common.EventHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, v := range e.handlers {
		if reflect.ValueOf(v).Pointer() == reflect.ValueOf(handler).Pointer() {
			return fmt.Errorf("duplicated")
		}
	}

	e.handlers = append(e.handlers, handler)
	return nil
}

func (e *event) Unsubscribe(handler common.EventHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i := 0; i < len(e.handlers); i++ {
		if reflect.ValueOf(e.handlers[i]).Pointer() == reflect.ValueOf(handler).Pointer() {
			e.handlers = append(e.handlers[:i], e.handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("not found")
}

func (e *event) Emit(args ...interface{}) []error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	res := []error{}

	for _, v := range e.handlers {
		if err := v(args...); err != nil {
			res = append(res, err)
		}
	}
	if len(res) > 0 {
		return res
	}
	return nil
}
