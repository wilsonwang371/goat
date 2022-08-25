package core

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"goat/pkg/config"
	"goat/pkg/db"
	"goat/pkg/logger"

	"go.uber.org/zap"
)

// interface for feed value generator
type FeedGenerator interface {
	PopNextValues() (time.Time, map[string]interface{}, Frequency, error)
	CreateDataSeries(key string, maxLen int) DataSeries
	PeekNextTime() *time.Time
	Finish()
	IsComplete() bool
	AppendNewValueToBuffer(time.Time, map[string]interface{}, Frequency) error
}

type barFeedGenerator struct {
	freq         []Frequency
	maxLen       int
	dataBuf      []*BarFeedGeneratorData
	dataBufMutex sync.Mutex
	eof          bool
}

// IsComplete implements FeedGenerator
func (b *barFeedGenerator) IsComplete() bool {
	return b.eof && len(b.dataBuf) == 0
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
// if we return no err,  still it can have no data returned
func (b *barFeedGenerator) PopNextValues() (time.Time, map[string]interface{}, Frequency, error) {
	b.dataBufMutex.Lock()
	defer b.dataBufMutex.Unlock()
	if len(b.dataBuf) == 0 {
		if b.eof {
			return time.Time{}, nil, 0, fmt.Errorf("feed generator is EOF")
		}
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
	if len(b.dataBuf) >= config.DataFeedMaxPendingBars {
		return fmt.Errorf("feed generator buffer is full")
	}
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
		dataBuf:      []*BarFeedGeneratorData{},
		dataBufMutex: sync.Mutex{},
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
	recoveryDB        *db.DB
	rows              *sql.Rows
}

// CreateDataSeries implements DataFeed
func (d *genericDataFeed) CreateDataSeries(key string, maxLen int) DataSeries {
	return d.feedGenerator.CreateDataSeries(key, maxLen)
}

// Dispatch implements DataFeed
func (d *genericDataFeed) Dispatch() bool {
	var t time.Time
	var v map[string]interface{}
	var f Frequency
	var err error

	if t, v, f, err = d.feedGenerator.PopNextValues(); err != nil {
		return false
	} else {
		// we may need to read data from recovery db
		if d.recoveryDB != nil {
			if d.rows.Next() {
				// NOTE: if we have data in the recovery database, we should use it
				var barData db.BarData
				if err := d.recoveryDB.ScanRows(d.rows, &barData); err != nil {
					logger.Logger.Error("failed to scan row", zap.Error(err))
					panic(err)
				}

				bar := NewBasicBar(time.Unix(barData.DateTime, 0),
					barData.Open,
					barData.High,
					barData.Low,
					barData.Close,
					barData.AdjClose,
					barData.Volume,
					Frequency(barData.Frequency))

				// NOTE:  we replace the value with the one from the recovery database
				t = time.Unix(barData.DateTime, 0)
				v = map[string]interface{}{}
				v[barData.Symbol] = bar.(interface{})
				f = Frequency(barData.Frequency)

			} else {
				d.rows.Close()
				d.recoveryDB = nil
			}
		}
	}

	if v != nil {
		if err := d.dataSeriesManager.newValueUpdate(t, v, f); err != nil {
			panic(err)
		}
		for key, val := range v {
			if b, ok := val.(Bar); ok {
				logger.Logger.Debug("dispatch a bar", zap.String("key", key), zap.Stringer("bar", b))
			}
		}
		// logger.Logger.Debug("emit new value", zap.Any("t", t), zap.Any("v", fmt.Sprintf("%+v", v)), zap.Any("f", f))
		d.newValueEvent.Emit(t, v)
		return true
	}
	return false
}

// Eof implements DataFeed
func (d *genericDataFeed) Eof() bool {
	// logger.Logger.Debug("feed eof", zap.Bool("eof", d.eof))
	return d.feedGenerator.IsComplete()
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
func NewGenericDataFeed(fg FeedGenerator, maxLen int, recoveryDB string) DataFeed {
	var recDB *db.DB
	var rows *sql.Rows
	if recoveryDB != "" {
		recDB = db.NewSQLiteDataBase(recoveryDB)
		if tmp, err := recDB.Model(&db.BarData{}).Order("primarykey").Rows(); err != nil {
			panic(err)
		} else {
			rows = tmp
		}
	}
	df := &genericDataFeed{
		newValueEvent: NewEvent(),
		feedGenerator: fg,
		recoveryDB:    recDB,
		rows:          rows,
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
	names := []string{}
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
