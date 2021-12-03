package feed_old

import (
	"fmt"
	"goalgotrade/common"
	"goalgotrade/core"
	lg "goalgotrade/logger"
	"time"

	"go.uber.org/zap"
)

type regDs struct {
	key  string
	freq common_old.Frequency
}

type BaseFeed struct {
	core.DefaultSubject
	event        common_old.Event
	dataSeries   map[string]map[common_old.Frequency]common_old.BarDataSeries
	registeredDs []regDs
	maxLen       int
}

func NewBaseFeed(maxLen int) *BaseFeed {
	subject := core.NewDefaultSubject()
	res := &BaseFeed{
		DefaultSubject: *subject,
		event:          core.NewEvent(),
		dataSeries:     map[string]map[common_old.Frequency]common_old.BarDataSeries{},
		maxLen:         maxLen,
	}
	res.Self = res
	return res
}

func (b *BaseFeed) GetMaxLen() int {
	return b.maxLen
}

func (b *BaseFeed) Reset() {
	b.dataSeries = make(map[string]map[common_old.Frequency]common_old.BarDataSeries)
	for _, v := range b.registeredDs {
		err := b.RegisterDataSeries(v.key, v.freq)
		if err != nil {
			lg.Logger.Warn("error", zap.Error(err))
		}
	}
}

func (b *BaseFeed) CreateDataSeries(key string, maxLen int) common_old.BarDataSeries {
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *BaseFeed) GetNextValues() (*time.Time, common_old.Bars, []common_old.Frequency, error) {
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *BaseFeed) GetNextValuesAndUpdateDS() (*time.Time, common_old.Bars, []common_old.Frequency, error) {
	dateTime, values, freqList, err := b.Self.(common_old.Feed).GetNextValues()
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
				b.dataSeries[k] = make(map[common_old.Frequency]common_old.BarDataSeries)
			} else {
				for _, freq := range freqList {
					if v2, ok2 := v[freq]; ok2 {
						for _, bar := range values.GetBarList(k) {
							if err := v2.Append(bar); err != nil {
								return nil, nil, nil, fmt.Errorf("error appeding bar")
							}
						}
					} else {
						b.dataSeries[k][freq] = b.Self.(common_old.Feed).CreateDataSeries(k, b.maxLen)
					}
				}
			}
		}
	}
	return dateTime, values, freqList, err
}

func (b *BaseFeed) RegisterDataSeries(key string, freq common_old.Frequency) error {
	if _, ok := b.dataSeries[key]; !ok {
		b.dataSeries[key] = map[common_old.Frequency]common_old.BarDataSeries{}
	}
	if _, ok := b.dataSeries[key][freq]; !ok {
		b.dataSeries[key][freq] = b.Self.(common_old.Feed).CreateDataSeries(key, b.maxLen)
		for _, v := range b.registeredDs {
			if v.key == key && v.freq == freq {
				return nil
			}
		}
		b.registeredDs = append(b.registeredDs, regDs{key: key, freq: freq})
	}
	return nil
}

func (b *BaseFeed) GetNewValuesEvent() common_old.Event {
	return b.event
}

func (b *BaseFeed) Dispatch() (bool, error) {
	// TODO: check if freq here is needed
	dateTime, values, _, err := b.Self.(common_old.Feed).GetNextValuesAndUpdateDS()
	if err != nil {
		lg.Logger.Debug("GetNextValuesAndUpdateDS failed", zap.Error(err))
		return false, err
	}
	if dateTime != nil {
		b.event.Emit(dateTime, values)
	}
	return dateTime != nil, nil
}

func (b *BaseFeed) Eof() bool {
	panic("not implemented")
}

func (b *BaseFeed) GetKeys() []string {
	var res []string
	for k := range b.dataSeries {
		res = append(res, k)
	}
	return res
}

func (b *BaseFeed) IsLive() bool {
	return false
}
