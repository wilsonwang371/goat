package core

import "time"

type DataSeries interface {
	AppendWithDateTime(timeVal time.Time, values interface{}) error
	Len() int
	Slice(start, end int) DataSeries
	Position(po int) (time.Time, interface{})
	DateTimes() []time.Time
}

type SequenceDataSeries interface {
	DataSeries
	GetDataSeriesNewValueEvent() Event
}

type sequenceDataSeries struct {
	event Event
}

// DataSeriesNewValueEvent implements SequenceDataSeries
func (d *sequenceDataSeries) GetDataSeriesNewValueEvent() Event {
	return d.event
}

// AppendWithDateTime implements DataSeries
func (s *sequenceDataSeries) AppendWithDateTime(timeVal time.Time, values interface{}) error {
	panic("unimplemented")
}

// DateTimes implements DataSeries
func (s *sequenceDataSeries) DateTimes() []time.Time {
	panic("unimplemented")
}

// Len implements DataSeries
func (s *sequenceDataSeries) Len() int {
	panic("unimplemented")
}

// Position implements DataSeries
func (s *sequenceDataSeries) Position(po int) (time.Time, interface{}) {
	panic("unimplemented")
}

// Slice implements DataSeries
func (s *sequenceDataSeries) Slice(start int, end int) DataSeries {
	panic("unimplemented")
}

func NewSequenceDataSeries(maxLen int) SequenceDataSeries {
	return &sequenceDataSeries{
		event: NewEvent(),
	}
}
