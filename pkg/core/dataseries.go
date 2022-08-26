package core

import (
	"encoding/json"
	"fmt"
	"time"
)

type DataSeries interface {
	AppendWithDateTime(timeVal time.Time, values interface{}) error
	Append(values interface{}) error
	Len() int
	Slice(start, end int) DataSeries
	At(po int) (time.Time, interface{}, error)
	DateTimes() []time.Time
	GetDataAsObjects(int) (map[string]interface{}, error)
}

type SequenceDataSeries interface {
	DataSeries
	GetDataSeriesNewValueEvent() Event
}

type sequenceDataSeries struct {
	event     Event
	maxLen    int
	dateTimes []time.Time
	values    []interface{}
}

type exportedObject struct {
	Data []interface{} `json:"data"`
}

// GetDataAsObjects implements DataSeries
func (s *sequenceDataSeries) GetDataAsObjects(length int) (map[string]interface{}, error) {
	if length <= 0 {
		return nil, fmt.Errorf("length should be greater than 0")
	}
	if length > s.Len() {
		length = s.Len()
	}
	rtn := exportedObject{
		Data: s.values[len(s.values)-length:],
	}
	// fmt.Printf("rtn: %+v\n", rtn)
	if rawData, err := json.Marshal(rtn); err == nil {
		var obj map[string]interface{}
		if err := json.Unmarshal(rawData, &obj); err == nil {
			return obj, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// Append implements DataSeries
func (s *sequenceDataSeries) Append(values interface{}) error {
	return s.AppendWithDateTime(time.Time{}, values)
}

// DataSeriesNewValueEvent implements SequenceDataSeries
func (d *sequenceDataSeries) GetDataSeriesNewValueEvent() Event {
	return d.event
}

// AppendWithDateTime implements DataSeries
func (s *sequenceDataSeries) AppendWithDateTime(timeVal time.Time, values interface{}) error {
	if !timeVal.IsZero() && len(s.dateTimes) != 0 && timeVal.Before(s.dateTimes[len(s.dateTimes)-1]) {
		return fmt.Errorf("new value before last value")
	}
	if len(s.dateTimes) != len(s.values) {
		panic("sequenceDataSeries.AppendWithDateTime: dateTimes and values are not of the same length")
	}
	if s.maxLen > 0 && len(s.values) >= s.maxLen {
		s.values = s.values[1:]
		s.dateTimes = s.dateTimes[1:]
	}
	s.values = append(s.values, values)
	s.dateTimes = append(s.dateTimes, timeVal)
	return nil
}

// DateTimes implements DataSeries
func (s *sequenceDataSeries) DateTimes() []time.Time {
	return s.dateTimes
}

// Len implements DataSeries
func (s *sequenceDataSeries) Len() int {
	if len(s.values) != len(s.dateTimes) {
		panic("len mismatch")
	}
	return len(s.values)
}

// At implements DataSeries
func (s *sequenceDataSeries) At(po int) (time.Time, interface{}, error) {
	if po < 0 || po >= len(s.values) {
		return time.Time{}, nil, fmt.Errorf("index out of range")
	}
	return s.dateTimes[po], s.values[po], nil
}

// Slice implements DataSeries
func (s *sequenceDataSeries) Slice(start int, end int) DataSeries {
	if start < 0 || start >= len(s.values) {
		panic("start index out of range")
	}
	if end < 0 || end >= len(s.values) {
		panic("end index out of range")
	}
	if start > end {
		panic("start index greater than end index")
	}
	return &sequenceDataSeries{
		event:     NewEvent(),
		maxLen:    s.maxLen,
		dateTimes: s.dateTimes[start:end],
		values:    s.values[start:end],
	}
}

func NewSequenceDataSeries(maxLen int) SequenceDataSeries {
	return &sequenceDataSeries{
		event:  NewEvent(),
		maxLen: maxLen,
	}
}

type BarDataSeries interface {
	SequenceDataSeries
	OpenDataSeries() DataSeries
	HighDataSeries() DataSeries
	LowDataSeries() DataSeries
	CloseDataSeries() DataSeries
	VolumeDataSeries() DataSeries
	AdjCloseDataSeries() DataSeries
}

type barDataSeries struct {
	sequenceDataSeries
	open         SequenceDataSeries
	high         SequenceDataSeries
	low          SequenceDataSeries
	close        SequenceDataSeries
	volume       SequenceDataSeries
	adjClose     SequenceDataSeries
	useAdjValues bool
	maxLen       int
}

// AdjCloseDataSeries implements BarDataSeries
func (b *barDataSeries) AdjCloseDataSeries() DataSeries {
	return b.adjClose
}

// CloseDataSeries implements BarDataSeries
func (b *barDataSeries) CloseDataSeries() DataSeries {
	return b.close
}

// HighDataSeries implements BarDataSeries
func (b *barDataSeries) HighDataSeries() DataSeries {
	return b.high
}

// LowDataSeries implements BarDataSeries
func (b *barDataSeries) LowDataSeries() DataSeries {
	return b.low
}

// OpenDataSeries implements BarDataSeries
func (b *barDataSeries) OpenDataSeries() DataSeries {
	return b.open
}

// VolumeDataSeries implements BarDataSeries
func (b *barDataSeries) VolumeDataSeries() DataSeries {
	return b.volume
}

// Append implements BarDataSeries
func (b *barDataSeries) Append(values interface{}) error {
	return b.AppendWithDateTime(time.Time{}, values)
}

// AppendWithDateTime implements BarDataSeries
func (b *barDataSeries) AppendWithDateTime(timeVal time.Time, values interface{}) error {
	bar := values.(Bar)
	bar.SetUseAdjustedValue(b.useAdjValues)
	b.sequenceDataSeries.AppendWithDateTime(timeVal, bar)

	b.open.AppendWithDateTime(timeVal, bar.Open())
	b.high.AppendWithDateTime(timeVal, bar.High())
	b.low.AppendWithDateTime(timeVal, bar.Low())
	b.close.AppendWithDateTime(timeVal, bar.Close())
	b.volume.AppendWithDateTime(timeVal, bar.Volume())
	b.adjClose.AppendWithDateTime(timeVal, bar.AdjClose())
	return nil
}

func NewBarDataSeries(maxLen int, useAdjValues bool) BarDataSeries {
	return &barDataSeries{
		sequenceDataSeries: sequenceDataSeries{
			event:  NewEvent(),
			maxLen: maxLen,
		},
		maxLen:       maxLen,
		useAdjValues: useAdjValues,
		open:         NewSequenceDataSeries(maxLen),
		high:         NewSequenceDataSeries(maxLen),
		low:          NewSequenceDataSeries(maxLen),
		close:        NewSequenceDataSeries(maxLen),
		volume:       NewSequenceDataSeries(maxLen),
		adjClose:     NewSequenceDataSeries(maxLen),
	}
}
