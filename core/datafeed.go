package core

import (
	"fmt"
	"time"
)

// interface for feed value generator
type FeedGenerator interface {
	NextValues() (time.Time, map[string]interface{}, Frequency, error)
	CreateDataSeries(key string, maxLen int) DataSeries
}

type barFeedGenerator struct {
	freq   []Frequency
	maxLen int
}

// CreateDataSeries implements FeedGenerator
func (*barFeedGenerator) CreateDataSeries(key string, maxLen int) DataSeries {
	panic("unimplemented")
}

// NextValues implements FeedGenerator
func (*barFeedGenerator) NextValues() (time.Time, map[string]interface{}, Frequency, error) {
	panic("unimplemented")
}

func NewBarFeedGenerator(freq []Frequency, maxLen int) FeedGenerator {
	return &barFeedGenerator{
		freq:   freq,
		maxLen: maxLen,
	}
}

type DataFeed interface {
	Subject
	GetNewValueEvent() Event
	CreateDataSeries(key string, maxLen int) DataSeries
}

type genericDataFeed struct {
	newValueEvent     Event
	dataSeriesManager *dataSeriesManager
	feedGenerator     FeedGenerator
}

// CreateDataSeries implements DataFeed
func (d *genericDataFeed) CreateDataSeries(key string, maxLen int) DataSeries {
	return d.feedGenerator.CreateDataSeries(key, maxLen)
}

// Dispatch implements DataFeed
func (d *genericDataFeed) Dispatch() bool {
	/*
			dateTime, values, _ = self.getNextValuesAndUpdateDS()
		        if dateTime is not None:
		            self.__event.emit(dateTime, values)
		        return dateTime is not None
	*/
	if t, v, f, err := d.feedGenerator.NextValues(); err != nil {
		return false
	} else {
		if err := d.dataSeriesManager.newValueUpdate(t, v, f); err != nil {
			panic(err)
		}
		d.newValueEvent.Emit(t, v)
	}
	return true
}

// Eof implements DataFeed
func (d *genericDataFeed) Eof() bool {
	return true
}

// Join implements DataFeed
func (d *genericDataFeed) Join() error {
	return nil
}

// PeekDateTime implements DataFeed
func (d *genericDataFeed) PeekDateTime() time.Time {
	return time.Now().UTC()
}

// Start implements DataFeed
func (d *genericDataFeed) Start() error {
	return nil
}

// Stop implements DataFeed
func (d *genericDataFeed) Stop() error {
	return nil
}

// GetNewValueEvent implements DataFeed
func (d *genericDataFeed) GetNewValueEvent() Event {
	return d.newValueEvent
}

// GetOrderUpdatedEvent implements Broker
func NewGenericDataFeed(fg FeedGenerator, maxLen int) DataFeed {
	df := &genericDataFeed{
		newValueEvent: NewEvent(),
		feedGenerator: fg,
	}
	df.dataSeriesManager = newDataSeriesManager(df, maxLen)
	return df
}

// internal data series manager
type dataSeriesManager struct {
	dataSeries map[string]DataSeries
	dataFeed   DataFeed
	maxLen     int
}

// crate new internal data series manager
func newDataSeriesManager(feed DataFeed, maxLen int) *dataSeriesManager {
	return &dataSeriesManager{
		dataSeries: make(map[string]DataSeries),
		dataFeed:   feed,
		maxLen:     maxLen,
	}
}

func (d *dataSeriesManager) registerDataSeries(name string, dataSeries DataSeries) error {
	if _, ok := d.dataSeries[name]; ok {
		return fmt.Errorf("data series %s already registered", name)
	}
	d.dataSeries[name] = dataSeries
	return nil
}

func (d *dataSeriesManager) getDataSeries(name string) (DataSeries, error) {
	if dataSeries, ok := d.dataSeries[name]; ok {
		return dataSeries, nil
	} else {
		return nil, fmt.Errorf("data series %s not found", name)
	}
}

func (d *dataSeriesManager) getDataSeriesNames() []string {
	names := make([]string, 0, len(d.dataSeries))
	for name := range d.dataSeries {
		names = append(names, name)
	}
	return names
}

func (d *dataSeriesManager) newValueUpdate(timeVal time.Time, values map[string]interface{}, freq Frequency) error {
	for key, value := range values {
		if dataSeries, err := d.getDataSeries(key); err == nil {
			if err := dataSeries.AppendWithDateTime(timeVal, value); err != nil {
				return err
			}
		} else {
			d.registerDataSeries(key, d.dataFeed.CreateDataSeries(key, d.maxLen))
		}
	}
	return nil
}
