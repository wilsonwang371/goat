package dataseries

import (
	"fmt"
	"goalgotrade/common"
	"goalgotrade/core"
	"sync"
	"time"
)

type SequenceDataSeries struct {
	Self          interface{}
	mu            sync.Mutex
	values        []interface{}
	dateTimes     []*time.Time
	newValueEvent common.Event
	maxLen        int
}

func NewSequenceDataSeries(maxLen int) *SequenceDataSeries {
	res := &SequenceDataSeries{
		newValueEvent: core.NewEvent(),
		maxLen:        maxLen,
	}
	res.Self = res
	return res
}

func (s *SequenceDataSeries) Len() int {
	return len(s.dateTimes)
}

func (s *SequenceDataSeries) MaxLen() int {
	return s.maxLen
}

func (s *SequenceDataSeries) SetMaxLen(maxLen int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxLen = maxLen
	if len(s.dateTimes) > s.maxLen {
		s.dateTimes = s.dateTimes[len(s.dateTimes)-s.maxLen:]
		s.values = s.values[len(s.values)-s.maxLen:]
	}
}

func (s *SequenceDataSeries) GetNewValueEvent() common.Event {
	return s.newValueEvent
}

func (s *SequenceDataSeries) Append(value interface{}) error {
	return s.AppendWithDateTime(nil, value)
}

func (s *SequenceDataSeries) AppendWithDateTime(dateTime *time.Time, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if dateTime != nil && len(s.dateTimes) > 0 && s.dateTimes[len(s.dateTimes)-1].After(*dateTime) {
		return fmt.Errorf("invalid datetime. it must be bigger than that last one")
	}
	s.dateTimes = append(s.dateTimes, dateTime)
	s.values = append(s.values, value)
	if len(s.dateTimes) > s.maxLen {
		s.dateTimes = s.dateTimes[len(s.dateTimes)-s.maxLen:]
		s.values = s.values[len(s.values)-s.maxLen:]
	}
	return nil
}

func (s *SequenceDataSeries) GetDateTimes() []*time.Time {
	return s.dateTimes
}
