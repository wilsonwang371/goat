package feed

import (
	"github.com/go-gota/gota/series"
	"goalgotrade/bar"
	"goalgotrade/consts/frequency"
	"time"
)

type fakeFetcherProvider struct {
	instrument       string
	freqList         []frequency.Frequency
}

func (f *fakeFetcherProvider) init(instrument string, freqList []frequency.Frequency) error {
	f.instrument = instrument
	f.freqList = freqList
	return nil
}

func (f *fakeFetcherProvider) connect() error {
	return nil
}

func (f *fakeFetcherProvider) nextBars() (bar.Bars, error) {
	bars := bar.NewBars()
	basicBar, err := bar.NewBasicBar(time.Now(), .0, .0, .0, .0, .0, .0, f.freqList[0])
	if err != nil {
		panic(err)
	}
	bars.AddBarList(f.instrument, []bar.Bar{basicBar})
	return bars, nil
}

func (f *fakeFetcherProvider) reset() error {
	return nil
}

func (f *fakeFetcherProvider) stop() error {
	return nil
}

func (f *fakeFetcherProvider) datatype() series.Type {
	return series.Float
}

func NewFakeFetcherProvider() BarFetcherProvider {
	return &fakeFetcherProvider{}
}
