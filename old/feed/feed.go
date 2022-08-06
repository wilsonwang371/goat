package feed

import (
	"fmt"
	"goalgotrade/bar"
	"goalgotrade/consts/frequency"
	"goalgotrade/core"
	"goalgotrade/dataseries"
	lg "goalgotrade/logger"
	"time"

	"go.uber.org/zap"
)

// BaseFeed ...
type BaseFeed interface {
	core.Subject
	Reset(f BaseFeed) error
	CreateDataSeries(key string, maxLen int) dataseries.DataSeries
	NextValues(f BaseFeed) (*time.Time, interface{}, []frequency.Frequency, error)
	RegisterDataSeries(f BaseFeed, key string, freq frequency.Frequency) error
	GetNextValuesAndUpdateDS(f BaseFeed) (*time.Time, interface{}, []frequency.Frequency, error)
	NewValueChannel() core.Channel
	Keys() []string
	Get(instrument string, freq frequency.Frequency) interface{}
}

type regDs struct {
	key  string
	freq frequency.Frequency
}

type baseFeed struct {
	newValueChannel core.Channel
	dataSeries      map[string]map[frequency.Frequency]dataseries.DataSeries
	registeredDs    []regDs
	maxLen          int
}

// Start ...
func (b *baseFeed) Start() error {
	lg.Logger.Debug("baseFeed Start() called")
	return nil
}

// Stop ...
func (b *baseFeed) Stop() error {
	panic("implement me")
}

// Join ...
func (b *baseFeed) Join() error {
	panic("implement me")
}

// Eof ...
func (b *baseFeed) Eof() bool {
	lg.Logger.Debug("baseFeed Eof() called")
	return true
}

// PeekDateTime ...
func (b *baseFeed) PeekDateTime() *time.Time {
	panic("implement me")
}

// GetDispatchPriority ...
func (b *baseFeed) GetDispatchPriority() int {
	return 0
}

// SetDispatchPriority ...
func (b *baseFeed) SetDispatchPriority(priority int) {
	panic("implement me")
}

// OnDispatcherRegistered ...
func (b *baseFeed) OnDispatcherRegistered(dispatcher core.Dispatcher) error {
	panic("implement me")
}

// CreateDataSeries ...
func (b *baseFeed) CreateDataSeries(key string, maxLen int) dataseries.DataSeries {
	panic("implement me")
}

// NextValues ...
func (b *baseFeed) NextValues(f BaseFeed) (*time.Time, interface{}, []frequency.Frequency, error) {
	panic("implement me")
}

// Get ...
func (b *baseFeed) Get(instrument string, freq frequency.Frequency) interface{} {
	panic("implement me")
}

// NewBaseFeed ...
func NewBaseFeed(maxLen int) BaseFeed {
	return newBaseFeed(maxLen)
}

func newBaseFeed(maxLen int) *baseFeed {
	return &baseFeed{
		dataSeries:      map[string]map[frequency.Frequency]dataseries.DataSeries{},
		maxLen:          maxLen,
		newValueChannel: core.NewChannel(),
	}
}

// Reset ...
func (b *baseFeed) Reset(f BaseFeed) error {
	b.dataSeries = map[string]map[frequency.Frequency]dataseries.DataSeries{}
	for _, v := range b.registeredDs {
		err := f.RegisterDataSeries(f, v.key, v.freq)
		if err != nil {
			lg.Logger.Warn("error", zap.Error(err))
			return err
		}
	}
	return nil
}

// RegisterDataSeries ...
func (b *baseFeed) RegisterDataSeries(f BaseFeed, key string, freq frequency.Frequency) error {
	if _, ok := b.dataSeries[key]; !ok {
		b.dataSeries[key] = map[frequency.Frequency]dataseries.DataSeries{}
	}
	if _, ok := b.dataSeries[key][freq]; !ok {
		b.dataSeries[key][freq] = f.CreateDataSeries(key, b.maxLen)
		for _, v := range b.registeredDs {
			if v.key == key && v.freq == freq {
				return nil
			}
		}
		b.registeredDs = append(b.registeredDs, regDs{key: key, freq: freq})
	}
	return nil
}

// NewValueChannel ...
func (b *baseFeed) NewValueChannel() core.Channel {
	return b.newValueChannel
}

// Keys ...
func (b *baseFeed) Keys() []string {
	var res []string
	for k := range b.dataSeries {
		res = append(res, k)
	}
	return res
}

// Dispatch ...
func (b *baseFeed) Dispatch(sub interface{}) (bool, error) {
	// TODO: check if freq here is needed
	f := sub.(BaseFeed)
	dsTime, values, _, err := f.GetNextValuesAndUpdateDS(f)
	if err != nil {
		lg.Logger.Debug("GetNextValuesAndUpdateDS failed", zap.Error(err))
		return false, err
	}
	if dsTime != nil {
		b.newValueChannel.Emit(core.NewBasicEvent("new_value", map[string]interface{}{
			"time": dsTime,
			"bars": values,
		}))
	}
	return dsTime != nil, nil
}

// GetNextValuesAndUpdateDS ...
func (b *baseFeed) GetNextValuesAndUpdateDS(f BaseFeed) (*time.Time, interface{}, []frequency.Frequency, error) {
	dateTime, nextValues, freqList, err := f.NextValues(f)
	if err == nil {
		if nextValues == nil {
			return nil, nil, nil, nil
		}
		// currently we dont support other than Bars.
		// TODO: add a more generic implementation for nextValues
		values := nextValues.(bar.Bars)
		keys := values.Instruments()
		if keys == nil || len(keys) == 0 {
			return nil, nil, nil, fmt.Errorf("no instruments found")
		}
		for _, k := range keys {
			if v, ok := b.dataSeries[k]; !ok {
				b.dataSeries[k] = make(map[frequency.Frequency]dataseries.DataSeries)
			} else {
				for _, freq := range freqList {
					if v2, ok2 := v[freq]; ok2 {
						bar := values.Bar(k)
						sequenceDS := v2.(dataseries.SequenceDataSeries)
						if err := sequenceDS.Append(bar); err != nil {
							return nil, nil, nil, fmt.Errorf("error appeding bar")
						}
					} else {
						b.dataSeries[k][freq] = f.CreateDataSeries(k, b.maxLen)
					}
				}
			}
		}
	}
	return dateTime, nextValues, freqList, err
}
