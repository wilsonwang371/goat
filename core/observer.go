package core

import (
	"goalgotrade/common"
	"time"
)

type DefaultSubject struct {
	dispatchPrio int
}

func NewDefaultSubject() *DefaultSubject {
	return &DefaultSubject{}
}

func (s *DefaultSubject) Start() error {
	return nil
}

func (s *DefaultSubject) Stop() error {
	return nil
}

func (s *DefaultSubject) Join() error {
	return nil
}

func (s *DefaultSubject) Eof() bool {
	panic("not implemented")
}

func (s *DefaultSubject) Dispatch() (bool, error) {
	return true, nil
}

func (s *DefaultSubject) PeekDateTime() *time.Time {
	return nil
}

func (s *DefaultSubject) GetDispatchPriority() int {
	return s.dispatchPrio
}

func (s *DefaultSubject) SetDispatchPriority(priority int) {
	s.dispatchPrio = priority
}

func (s *DefaultSubject) OnDispatcherRegistered(dispatcher common.Dispatcher) error {
	return nil
}
