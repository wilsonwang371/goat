package barfeed

import (
	"fmt"
	"goalgotrade/common"
	"goalgotrade/dataseries"
	"goalgotrade/feed"
	lg "goalgotrade/logger"
	"time"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"
)

type baseBarFeed struct {
	*feed.BaseFeed
	frequencies       []common.Frequency
	useAdjustedValue  bool
	stype             series.Type
	defaultInstrument string
	currentBars       common.Bars
	lastBars          map[string]common.Bar
}

func NewBaseBarFeed(frequencies []common.Frequency, stype series.Type, maxlen int) common.BarFeed {
	return &baseBarFeed{
		BaseFeed:         feed.NewBaseFeed(maxlen),
		frequencies:      frequencies,
		useAdjustedValue: false,
		stype:            stype,
		lastBars:         map[string]common.Bar{},
	}
}

func (b *baseBarFeed) GetCurrentBars() common.Bars {
	return b.currentBars
}

func (b *baseBarFeed) GetLastBar(instrument string) common.Bar {
	if v, ok := b.lastBars[instrument]; ok {
		return v
	}
	return nil
}

func (b *baseBarFeed) GetNextBars() common.Bars {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) GetNextValues() (*time.Time, common.Bars, []common.Frequency, error) {
	bars := b.GetNextBars()
	if bars == nil {
		freqs := bars.GetFrequencies()
		datetime := bars.GetDateTime()

		if len(freqs) == 0 || datetime == nil {
			lg.Logger.Error("invalid frequency and/or datetime", zap.Any("Frequencies", freqs), zap.Time("DateTime", *datetime))
			return nil, nil, []common.Frequency{}, fmt.Errorf("invalid frequency and/or datetime")
		}

		if b.currentBars != nil && b.currentBars.GetDateTime().After(*datetime) {
			return nil, nil, []common.Frequency{},
				fmt.Errorf("Bar date times are not in order. Previous datetime was %s and current datetime is %s",
					b.currentBars.GetDateTime(), datetime)
		}

		b.currentBars = bars
		for _, v := range bars.GetInstruments() {
			b.lastBars[v] = bars.GetBar(v)
		}
		return datetime, bars, freqs, nil
	}
	return nil, nil, []common.Frequency{}, fmt.Errorf("no next bars")
}

func (b *baseBarFeed) GetCurrentDateTime() *time.Time {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) BarsHaveAdjClose() bool {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) GetFrequencies() []common.Frequency {
	return b.frequencies
}

func (b *baseBarFeed) CreateDataSeries(key string, maxlen int) common.BarDataSeries {
	ret := dataseries.NewBarDataSeries(b.stype, b.BaseFeed.GetMaxLen())
	return ret
}

func (b *baseBarFeed) GetDefaultInstrument() string {
	return b.defaultInstrument
}

func (b *baseBarFeed) GetRegisteredInstruments() []string {
	return b.GetKeys()
}

func (b *baseBarFeed) RegisterInstrument(instrument string, freq common.Frequency) error {
	b.defaultInstrument = instrument
	err := b.RegisterDataSeries(instrument, freq)
	return err
}

func (b *baseBarFeed) GetDataSeries(instrument string, freq common.Frequency) *series.Series {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}
