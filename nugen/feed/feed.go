package feed

import (
	"fmt"
	"goalgotrade/nugen/bar"
	"goalgotrade/nugen/consts/frequency"
	"goalgotrade/nugen/core"
	"goalgotrade/nugen/dataseries"
	lg "goalgotrade/nugen/logger"
	"time"

	"go.uber.org/zap"
)

// BaseFeed ...
type BaseFeed interface {
	core.Subject
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

type partialBaseFeed struct {
	newValueChannel core.Channel
	dataSeries      map[string]map[frequency.Frequency]dataseries.DataSeries
	registeredDs    []regDs
	maxLen          int
}

func newPartialBaseFeed(maxLen int) *partialBaseFeed {
	return &partialBaseFeed{
		dataSeries:      map[string]map[frequency.Frequency]dataseries.DataSeries{},
		maxLen:          maxLen,
		newValueChannel: core.NewChannel(),
	}
}

// Reset ...
func (b *partialBaseFeed) Reset(f BaseFeed) error {
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
func (b *partialBaseFeed) RegisterDataSeries(f BaseFeed, key string, freq frequency.Frequency) error {
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
func (b *partialBaseFeed) NewValueChannel() core.Channel {
	return b.newValueChannel
}

// Keys ...
func (b *partialBaseFeed) Keys() []string {
	var res []string
	for k := range b.dataSeries {
		res = append(res, k)
	}
	return res
}

// Dispatch ...
func (b *partialBaseFeed) Dispatch(sub interface{}) (bool, error) {
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
func (b *partialBaseFeed) GetNextValuesAndUpdateDS(f BaseFeed) (*time.Time, interface{}, []frequency.Frequency, error) {
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
						for _, bar := range values.Items(k) {
							sequenceDS := v2.(dataseries.SequenceDataSeries)
							if err := sequenceDS.Append(bar); err != nil {
								return nil, nil, nil, fmt.Errorf("error appeding bar")
							}
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
