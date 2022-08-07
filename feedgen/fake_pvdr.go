package feedgen

import (
	"goalgotrade/core"
	"time"

	"github.com/go-gota/gota/series"
)

type fakeDataProvider struct {
	instrument string
	freqList   []core.Frequency
}

func (f *fakeDataProvider) init(instrument string, freqList []core.Frequency) error {
	f.instrument = instrument
	f.freqList = freqList
	return nil
}

func (f *fakeDataProvider) connect() error {
	return nil
}

func (f *fakeDataProvider) nextBars() (map[string]core.Bar, error) {
	basicBar := core.NewBasicBar(.1, .2, .3, .4, 5, f.freqList[0], time.Now())
	time.Sleep(time.Second)
	res := make(map[string]core.Bar)
	res[f.instrument] = basicBar
	return res, nil
}

func (f *fakeDataProvider) reset() error {
	return nil
}

func (f *fakeDataProvider) stop() error {
	return nil
}

func (f *fakeDataProvider) datatype() series.Type {
	return series.Float
}

func NewFakeDataProvider() BarDataProvider {
	return &fakeDataProvider{}
}
