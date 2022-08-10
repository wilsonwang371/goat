package core

import (
	"fmt"
	"goalgotrade/pkg/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

// interface for feed value generator
type FeedGenerator interface {
	PopNextValues() (time.Time, map[string]interface{}, Frequency, error)
	CreateDataSeries(key string, maxLen int) DataSeries
	PeekNextTime() *time.Time
	Finish()
	AppendNewValueToBuffer(time.Time, map[string]interface{}, Frequency) error
}

type barFeedGenerator struct {
	freq         []Frequency
	maxLen       int
	dataBuf      []*BarFeedGeneratorData
	dataBufMutex sync.Mutex
	eof          bool
}

// PeekNextTime implements FeedGenerator
func (b *barFeedGenerator) PeekNextTime() *time.Time {
	b.dataBufMutex.Lock()
	defer b.dataBufMutex.Unlock()
	if len(b.dataBuf) == 0 {
		return nil
	}
	elem := b.dataBuf[0]
	t := elem.t
	return &t
}

type BarFeedGeneratorData struct {
	t time.Time
	d map[string]interface{}
	f Frequency
}

// CreateDataSeries implements FeedGenerator
func (b *barFeedGenerator) CreateDataSeries(key string, maxLen int) DataSeries {
	bds := NewBarDataSeries(maxLen, false)
	return bds
}

// NextValues implements FeedGenerator
// NOTE: we cannot simply wait here because it is called inside dispatch loop
func (b *barFeedGenerator) PopNextValues() (time.Time, map[string]interface{}, Frequency, error) {
	b.dataBufMutex.Lock()
	defer b.dataBufMutex.Unlock()
	if b.eof {
		return time.Time{}, nil, 0, fmt.Errorf("feed generator is EOF")
	}
	if len(b.dataBuf) == 0 {
		return time.Time{}, nil, 0, nil
	}
	elem := b.dataBuf[0]
	b.dataBuf = b.dataBuf[1:]
	return elem.t, elem.d, elem.f, nil
}

func (b *barFeedGenerator) Finish() {
	b.eof = true
}

func (b *barFeedGenerator) AppendNewValueToBuffer(t time.Time, v map[string]interface{}, f Frequency) error {
	found := false
	for i := 0; i < len(b.freq); i++ {
		if b.freq[i] == f {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("frequency %d not valid for generator", f)
	}
	if b.eof {
		return fmt.Errorf("feed generator is closed")
	}

	b.dataBufMutex.Lock()
	defer b.dataBufMutex.Unlock()
	b.dataBuf = append(b.dataBuf, &BarFeedGeneratorData{
		t: t,
		d: v,
		f: f,
	})
	return nil
}

func NewBarFeedGenerator(freq []Frequency, maxLen int) FeedGenerator {
	return &barFeedGenerator{
		freq:         freq,
		maxLen:       maxLen,
		dataBuf:      make([]*BarFeedGeneratorData, 0),
		dataBufMutex: sync.Mutex{},
		eof:          false,
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
	eof               bool
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
	if t, v, f, err := d.feedGenerator.PopNextValues(); err != nil {
		d.eof = true
		return false
	} else if v != nil {
		if err := d.dataSeriesManager.newValueUpdate(t, v, f); err != nil {
			panic(err)
		}
		for key, val := range v {
			if b, ok := val.(Bar); ok {
				logger.Logger.Debug("Dispatch a bar", zap.String("key", key), zap.Stringer("bar", b))
			}
		}
		logger.Logger.Debug("emit new value", zap.Any("t", t), zap.Any("v", v), zap.Any("f", f))
		d.newValueEvent.Emit(t, v)
		return true
	}
	return false
}

// Eof implements DataFeed
func (d *genericDataFeed) Eof() bool {
	return d.eof
}

// Join implements DataFeed
func (d *genericDataFeed) Join() error {
	return nil
}

// PeekDateTime implements DataFeed
func (d *genericDataFeed) PeekDateTime() *time.Time {
	// NOTE: we need to read data
	return d.feedGenerator.PeekNextTime()
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
	logger.Logger.Info("GetNewValueEvent")
	return d.newValueEvent
}

// GetOrderUpdatedEvent implements Broker
func NewGenericDataFeed(fg FeedGenerator, maxLen int) DataFeed {
	df := &genericDataFeed{
		newValueEvent: NewEvent(),
		feedGenerator: fg,
		eof:           false,
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