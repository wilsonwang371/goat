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

type regDs struct {
	key  string
	freq frequency.Frequency
}

type InheritedFeedOps interface {
	CreateDataSeries(key string, maxLen int) dataseries.DataSeries
	RegisterDataSeries(f InheritedFeedOps, key string, freq frequency.Frequency) error
	NextValues(f InheritedFeedOps) (*time.Time, *bar.Bars, []frequency.Frequency, error)
	GetNextValuesAndUpdateDS(f InheritedFeedOps) (*time.Time, *bar.Bars, []frequency.Frequency, error)
}

type Feed interface {
	core.Subject
	InheritedFeedOps
	NewValueChannel() *core.Channel
	MaxLen() int
	Keys() []string
}

func NewBaseFeed(maxLen int) *BaseFeed {
	return &BaseFeed{
		dataSeries:      map[string]map[frequency.Frequency]dataseries.DataSeries{},
		maxLen:          maxLen,
		newValueChannel: core.NewChannel(),
	}
}

type BaseFeed struct {
	newValueChannel core.Channel
	dataSeries      map[string]map[frequency.Frequency]dataseries.DataSeries
	registeredDs    []regDs
	maxLen          int
}

func (b *BaseFeed) Reset(f Feed) error {
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

func (b *BaseFeed) RegisterDataSeries(f InheritedFeedOps, key string, freq frequency.Frequency) error {
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

func (b *BaseFeed) GetNextValuesAndUpdateDS(f InheritedFeedOps) (*time.Time, *bar.Bars, []frequency.Frequency, error) {
	dateTime, values, freqList, err := f.NextValues(f)
	if err == nil {
		if values == nil {
			return nil, nil, nil, nil
		}
		keys := values.GetInstruments()
		if keys == nil || len(keys) == 0 {
			return nil, nil, nil, fmt.Errorf("no instruments found")
		}
		for _, k := range keys {
			if v, ok := b.dataSeries[k]; !ok {
				b.dataSeries[k] = make(map[frequency.Frequency]dataseries.DataSeries)
			} else {
				for _, freq := range freqList {
					if v2, ok2 := v[freq]; ok2 {
						for _, bar := range values.GetBarList(k) {
							sequenceDS := v2.(*dataseries.SequenceDataSeries)
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
	return dateTime, values, freqList, err
}

func (b *BaseFeed) NewValueChannel() *core.Channel {
	return b.newValueChannel
}

func (b *BaseFeed) MaxLen() int {
	return b.maxLen
}

func (b *BaseFeed) Keys() []string {
	var res []string
	for k := range b.dataSeries {
		res = append(res, k)
	}
	return res
}

func (b *BaseFeed) Dispatch(f Feed) (bool, error) {
	// TODO: check if freq here is needed
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
