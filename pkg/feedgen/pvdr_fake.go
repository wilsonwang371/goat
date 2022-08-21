package feedgen

import (
	"fmt"
	"time"

	"goat/pkg/core"

	"github.com/go-gota/gota/series"
)

type fakeDataProvider struct {
	instrument string
	freqList   []core.Frequency
	stopped    bool
}

func (f *fakeDataProvider) init(instrument string, freqList []core.Frequency) error {
	f.instrument = instrument
	f.freqList = freqList
	return nil
}

func (f *fakeDataProvider) connect() error {
	return nil
}

func (f *fakeDataProvider) nextBars() (core.Bars, error) {
	// this can return nothing but with no error, you should not block this forever
	if f.stopped {
		return nil, fmt.Errorf("fake data provider is stopped")
	}
	basicBar := core.NewBasicBar(time.Now(), .1, .2, .3, .4, .4, 5, f.freqList[0])
	time.Sleep(time.Second)
	res := make(core.Bars)
	res[f.instrument] = basicBar
	return res, nil
}

func (f *fakeDataProvider) reset() error {
	return nil
}

func (f *fakeDataProvider) stop() error {
	f.stopped = true
	return nil
}

func (f *fakeDataProvider) datatype() series.Type {
	return series.Float
}

func NewFakeDataProvider() BarDataProvider {
	return &fakeDataProvider{
		stopped: false,
	}
}
