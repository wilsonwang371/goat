package barfeed

import (
	"fmt"
	"goalgotrade/nugen/bar"
	"goalgotrade/nugen/consts/frequency"
	"goalgotrade/nugen/dataseries"
	"goalgotrade/nugen/feed"
	lg "goalgotrade/nugen/logger"
	"time"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"
)

type BarFeed interface {
	feed.Feed
	NextBars() (*bar.Bars, error)
}

type BaseBarFeed struct {
	feed.BaseFeed
	frequencies       []frequency.Frequency
	useAdjustedValue  bool
	sType             series.Type
	defaultInstrument string
	currentBars       *bar.Bars
	lastBars          map[string][]*bar.BasicBar
}

func (b *BaseBarFeed) CreateDataSeries(key string, maxLen int) dataseries.DataSeries {
	// TODO: implement me and confirm if this is correct usage of dataseries
	ret := dataseries.NewBarDataSeries(b.sType, b.BaseFeed.MaxLen())
	return ret
}

func (b *BaseBarFeed) NextValues(f feed.InheritedFeedOps) (*time.Time, *bar.Bars, []frequency.Frequency, error) {
	bf := f.(BarFeed)
	bars, err := bf.NextBars()
	if bars != nil && err == nil {
		freqList := bars.GetFrequencies()
		barsTime := bars.Time()

		if len(freqList) == 0 || barsTime == nil {
			lg.Logger.Error("invalid frequency and/or dateTime", zap.Any("Frequencies", freqList), zap.Any("DateTime", barsTime))
			return nil, nil, []frequency.Frequency{}, fmt.Errorf("invalid frequency and/or dateTime")
		}

		if b.currentBars != nil && b.currentBars.Time().After(*barsTime) {
			return nil, nil, []frequency.Frequency{},
				fmt.Errorf("bar date times are not in order. Previous dateTime was %s and current dateTime is %s",
					b.currentBars.Time(), barsTime)
		}

		b.currentBars = bars
		for _, v := range bars.GetInstruments() {
			b.lastBars[v] = bars.GetBarList(v)
		}
		return barsTime, bars, freqList, nil
	}
	// it is okay to return nil with no error. it is idle() case
	return nil, nil, []frequency.Frequency{}, err
}

func NewBaseBarFeed(frequencies []frequency.Frequency, sType series.Type, maxLen int) *BaseBarFeed {
	baseFeed := feed.NewBaseFeed(maxLen)
	res := &BaseBarFeed{
		BaseFeed:         *baseFeed,
		frequencies:      frequencies,
		useAdjustedValue: false,
		sType:            sType,
		lastBars:         map[string][]*bar.BasicBar{},
	}
	return res
}

func (b *BaseBarFeed) Reset() {
	b.currentBars = nil
	b.lastBars = map[string][]*bar.BasicBar{}
}

func (b *BaseBarFeed) GetCurrentBars() *bar.Bars {
	return b.currentBars
}

func (b *BaseBarFeed) GetLastBar(instrument string) []*bar.BasicBar {
	if v, ok := b.lastBars[instrument]; ok {
		return v
	}
	return nil
}

func (b *BaseBarFeed) GetNextBars() (*bar.Bars, error) {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *BaseBarFeed) GetCurrentDateTime() *time.Time {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *BaseBarFeed) BarsHaveAdjClose() bool {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *BaseBarFeed) GetFrequencies() []frequency.Frequency {
	return b.frequencies
}

func (b *BaseBarFeed) GetDefaultInstrument() string {
	return b.defaultInstrument
}

func (b *BaseBarFeed) GetRegisteredInstruments() []string {
	return b.Keys()
}

func (b *BaseBarFeed) RegisterInstrument(instrument string, freqList []frequency.Frequency) error {
	b.defaultInstrument = instrument
	for _, freq := range freqList {
		err := b.RegisterDataSeries(b, instrument, freq)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BaseBarFeed) GetDataSeries(instrument string, freq frequency.Frequency) *series.Series {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}
