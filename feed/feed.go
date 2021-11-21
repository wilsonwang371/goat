package feed

import (
	"goalgotrade/common"
	"goalgotrade/core"
	lg "goalgotrade/logger"
	"time"
)

type regDs struct {
	key  string
	freq common.Frequency
}

type BaseFeed struct {
	core.DefaultSubject
	event        common.Event
	dataseries   map[string]map[common.Frequency]common.BarDataSeries
	registeredDs []regDs
	maxlen       int
}

func NewBaseFeed(maxlen int) *BaseFeed {
	subject := core.NewDefaultSubject()
	return &BaseFeed{
		DefaultSubject: *subject,
		event:          core.NewEvent(),
		dataseries:     map[string]map[common.Frequency]common.BarDataSeries{},
		maxlen:         maxlen,
	}
}

func (b *BaseFeed) GetMaxLen() int {
	return b.maxlen
}

func (b *BaseFeed) Reset() {
	b.dataseries = make(map[string]map[common.Frequency]common.BarDataSeries)
	for _, v := range b.registeredDs {
		b.RegisterDataSeries(v.key, v.freq)
	}
}

func (b *BaseFeed) CreateDataSeries(key string, maxlen int) common.BarDataSeries {
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *BaseFeed) GetNextValues() (*time.Time, common.Bars, []common.Frequency, error) {
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *BaseFeed) GetNextValuesAndUpdateDS() (*time.Time, common.Bars, []common.Frequency, error) {
	dateTime, values, freqs, err := interface{}(b).(common.Feed).GetNextValues()
	if err != nil || dateTime == nil {
		keys := values.GetInstruments()
		for _, k := range keys {
			if v, ok := b.dataseries[k]; !ok {
				b.dataseries[k] = make(map[common.Frequency]common.BarDataSeries)
			} else {
				for _, freq := range freqs {
					if v2, ok2 := v[freq]; ok2 {
						for _, bar := range values.GetBarList(k) {
							v2.Append(bar)
						}
					} else {
						b.dataseries[k][freq] = b.CreateDataSeries(k, b.maxlen)
					}
				}
			}
		}
	}
	return dateTime, values, freqs, err
}

func (b *BaseFeed) RegisterDataSeries(key string, freq common.Frequency) error {
	if _, ok := b.dataseries[key]; !ok {
		b.dataseries[key] = map[common.Frequency]common.BarDataSeries{}
	}
	if _, ok := b.dataseries[key][freq]; !ok {
		b.dataseries[key][freq] = b.CreateDataSeries(key, b.maxlen)
		for _, v := range b.registeredDs {
			if v.key == key && v.freq == freq {
				return nil
			}
		}
		b.registeredDs = append(b.registeredDs, regDs{key: key, freq: freq})
	}
	return nil
}

func (b *BaseFeed) GetNewValuesEvent() common.Event {
	return b.event
}

func (b *BaseFeed) Dispatch() (bool, error) {
	// TODO: check if freq here is needed
	dateTime, values, _, err := interface{}(b).(common.Feed).GetNextValuesAndUpdateDS()
	if err != nil {
		return false, err
	}
	if dateTime != nil {
		b.event.Emit(dateTime, values)
	}
	return dateTime != nil && err == nil, nil
}

func (b *BaseFeed) Eof() bool {
	panic("not implemented")
}

func (b *BaseFeed) GetKeys() []string {
	res := []string{}
	for k := range b.dataseries {
		res = append(res, k)
	}
	return res
}
