package core

import (
	"fmt"
	"sync"
	"time"

	"goat/pkg/config"
	"goat/pkg/db"
	"goat/pkg/logger"

	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
)

const FailOnUnmatchedSymbol = true

var warnedUnmatchedSymbol = false

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
	GetDataSeries(string, Frequency) (DataSeries, error)
}

type PendingDataFeedValue struct {
	t time.Time
	v map[string]interface{}
	f Frequency
}

type genericDataFeed struct {
	cfg               *config.Config
	newValueEvent     Event
	dataSeriesManager *dataSeriesManager
	feedGenerator     FeedGenerator

	recoveryDB       *db.DB
	recoveryCount    int64
	recoveryProgress int64
	recoveryBar      *progressbar.ProgressBar
	recoveryLastTime time.Time

	pendingRecoveryData  []*PendingDataFeedValue
	dataFeedHooksControl DataFeedHooksControl
}

// GetDataSeries implements DataFeed
func (d *genericDataFeed) GetDataSeries(symbol string, freq Frequency) (DataSeries, error) {
	return d.dataSeriesManager.getDataSeries(symbol, freq)
}

func (d *genericDataFeed) maybeFetchNextGeneratedData() {
	if v := d.dataFeedHooksControl.PossibleOneNewValue(); v != nil {
		d.pendingRecoveryData = append(d.pendingRecoveryData, v)
	}
}

func (d *genericDataFeed) maybeFetchNextRecoveryData() {
	var t time.Time
	var v map[string]interface{}
	var f Frequency

	if len(d.pendingRecoveryData) != 0 {
		// we have pending data, just ignore
		return
	}

	if d.recoveryDB != nil {
		if barData, err := d.recoveryDB.Next(); err == nil {
			// NOTE: if we have data in the recovery database, we should use it
			if barData == nil {
				// no more data in the recovery database
				d.recoveryDB = nil
				d.recoveryBar.Finish()
				logger.Logger.Info("recovery database is finished")
				return
			}

			// update progress bar
			d.recoveryProgress++
			if d.recoveryLastTime.IsZero() || time.Now().Sub(d.recoveryLastTime) > 10*time.Second {
				d.recoveryBar.Set64(d.recoveryProgress)
				d.recoveryLastTime = time.Now()
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
			if barData.Symbol != d.cfg.Symbol && !warnedUnmatchedSymbol {
				if FailOnUnmatchedSymbol {
					panic(fmt.Errorf("unmatched symbol %s in recovery database, expected %s",
						barData.Symbol, d.cfg.Symbol))
				}
				logger.Logger.Info("symbol in recovery database is different from config",
					zap.String("symbol", barData.Symbol),
					zap.String("config_symbol", d.cfg.Symbol))
				warnedUnmatchedSymbol = true
			}
			v[barData.Symbol] = bar.(interface{})
			f = Frequency(barData.Frequency)

			// logger.Logger.Info("read bar from recovery database",
			//	zap.Time("time", t), zap.String("symbol", barData.Symbol), zap.Any("bar", bar))
			d.pendingRecoveryData = append(d.pendingRecoveryData, &PendingDataFeedValue{t, v, f})
			return
		} else {
			// error happens and we cannot continue
			d.recoveryDB = nil
			panic(err)
		}
	} else {
		// no recovery database, just fetch from feed generator
		return
	}
}

// Dispatch implements DataFeed
func (d *genericDataFeed) Dispatch() bool {
	var t time.Time
	var v map[string]interface{}
	var f Frequency
	var err error
	var isRecovery bool

	if t, v, f, err = d.feedGenerator.PopNextValues(); err != nil {
		return false
	}

	// we may need to read data from recovery db
	d.maybeFetchNextRecoveryData()
	// we may need to read data from feed hooks
	d.maybeFetchNextGeneratedData()

	if d.recoveryDB != nil {
		isRecovery = true
	} else {
		isRecovery = false
	}

	if len(d.pendingRecoveryData) != 0 {
		// we have data from the recovery database, use it
		pendingData := d.pendingRecoveryData[0]
		t = pendingData.t
		v = pendingData.v
		f = pendingData.f
		d.pendingRecoveryData = d.pendingRecoveryData[1:]
	}

	if v != nil {
		d.dataFeedHooksControl.FilterNewValue(&PendingDataFeedValue{t, v, f}, isRecovery)

		if err := d.dataSeriesManager.newValueUpdate(t, v, f); err != nil {
			panic(err)
		}
		for _, val := range v {
			if _, ok := val.(Bar); !ok {
				panic(fmt.Errorf("value is not a bar"))
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
	d.maybeFetchNextRecoveryData()
	if len(d.pendingRecoveryData) != 0 {
		// we have data from the recovery database, use it
		return &d.pendingRecoveryData[0].t
	}
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
	// logger.Logger.Info("GetNewValueEvent")
	return d.newValueEvent
}

// GetOrderUpdatedEvent implements Broker
func NewGenericDataFeed(cfg *config.Config, fg FeedGenerator, hooksCtrl DataFeedHooksControl, maxLen int, recoveryDB string) DataFeed {
	var recDB *db.DB
	var recCount int64
	if recoveryDB != "" {
		logger.Logger.Debug("recovery mode is enabled", zap.String("db", recoveryDB))
		recDB = db.NewSQLiteDataBase(recoveryDB, false)
		recCount = recDB.FetchAll(true)
	}
	if hooksCtrl == nil {
		// default hooks
		hooksCtrl = NewDataFeedValueHookControl()
		hooksCtrl.AddNewHook(NewDayBarGenHook())
	}
	df := &genericDataFeed{
		cfg:                  cfg,
		newValueEvent:        NewEvent(),
		feedGenerator:        fg,
		recoveryDB:           recDB,
		recoveryCount:        recCount,
		recoveryProgress:     0,
		recoveryBar:          progressbar.Default(recCount),
		pendingRecoveryData:  []*PendingDataFeedValue{},
		dataFeedHooksControl: hooksCtrl,
	}
	df.dataSeriesManager = newDataSeriesManager(fg.CreateDataSeries, maxLen)
	return df
}

// internal data series manager
type dataSeriesManager struct {
	dataSeries   map[string]map[Frequency]DataSeries
	createDSFunc func(key string, maxLen int) DataSeries
	maxLen       int
}

// crate new internal data series manager
func newDataSeriesManager(f func(string, int) DataSeries, maxLen int) *dataSeriesManager {
	return &dataSeriesManager{
		dataSeries:   make(map[string]map[Frequency]DataSeries),
		createDSFunc: f,
		maxLen:       maxLen,
	}
}

func (d *dataSeriesManager) registerDataSeries(name string, freq Frequency, dataSeries DataSeries) error {
	if v, ok := d.dataSeries[name]; ok {
		if _, ok2 := v[freq]; ok2 {
			return fmt.Errorf("data series %s already registered for frequency %d", name, freq)
		}
		v[freq] = dataSeries
	} else {
		d.dataSeries[name] = make(map[Frequency]DataSeries)
		d.dataSeries[name][freq] = dataSeries
	}
	return nil
}

func (d *dataSeriesManager) getDataSeries(name string, freq Frequency) (DataSeries, error) {
	if dataSeries, ok := d.dataSeries[name]; ok {
		if v, ok := dataSeries[freq]; ok {
			return v, nil
		}
		return nil, fmt.Errorf("data series %s not registered for frequency %d", name, freq)
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
		if dataSeries, err := d.getDataSeries(key, freq); err == nil {
			if err := dataSeries.AppendWithDateTime(timeVal, value); err != nil {
				return err
			}
		} else {
			d.registerDataSeries(key, freq, d.createDSFunc(key, d.maxLen))
		}
	}
	return nil
}
