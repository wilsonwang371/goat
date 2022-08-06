package dataseries

import (
	"fmt"
	"goalgotrade/core"
	"sync"
	"time"
)

// SequenceDataSeries ...
type SequenceDataSeries interface {
	DataSeries
	MaxLen() int
	SetMaxLen(maxLen int)
	NewValueChannel() core.Channel
	Append(value interface{}) error
	AppendWithDateTime(dateTime *time.Time, value interface{}) error
}

type sequenceDataSeries struct {
	mu              sync.Mutex
	values          []interface{}
	times           []*time.Time
	newValueChannel core.Channel
	maxLen          int
}

// NewSequenceDataSeries ...
func NewSequenceDataSeries(maxLen int) SequenceDataSeries {
	res := &sequenceDataSeries{
		newValueChannel: core.NewChannel(),
		maxLen:          maxLen,
	}
	return res
}

// Len ...
func (s *sequenceDataSeries) Len() int {
	return len(s.times)
}

// MaxLen ...
func (s *sequenceDataSeries) MaxLen() int {
	return s.maxLen
}

// SetMaxLen ...
func (s *sequenceDataSeries) SetMaxLen(maxLen int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxLen = maxLen
	if len(s.times) > s.maxLen {
		s.times = s.times[len(s.times)-s.maxLen:]
		s.values = s.values[len(s.values)-s.maxLen:]
	}
}

// NewValueChannel ...
func (s *sequenceDataSeries) NewValueChannel() core.Channel {
	return s.newValueChannel
}

// AtIndex ...
func (s *sequenceDataSeries) AtIndex(index int) interface{} {
	if index < 0 || index >= len(s.values) {
		return nil
	}
	return s.values[index]
}

// Append ...
func (s *sequenceDataSeries) Append(value interface{}) error {
	return s.AppendWithDateTime(nil, value)
}

// AppendWithDateTime ...
func (s *sequenceDataSeries) AppendWithDateTime(dateTime *time.Time, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if dateTime != nil && len(s.times) > 0 && s.times[len(s.times)-1].After(*dateTime) {
		return fmt.Errorf("invalid datetime. it must be bigger than that last one")
	}
	s.times = append(s.times, dateTime)
	s.values = append(s.values, value)
	if len(s.times) > s.maxLen {
		s.times = s.times[len(s.times)-s.maxLen:]
		s.values = s.values[len(s.values)-s.maxLen:]
	}
	return nil
}

// Times ...
func (s *sequenceDataSeries) Times() []*time.Time {
	return s.times
}

// Get ...
func (s *sequenceDataSeries) Get(key string) (interface{}, bool) {
	panic("not implemented")
}

// Set ...
func (s *sequenceDataSeries) Set(key string, value interface{}) error {
	panic("not implemented")
}
