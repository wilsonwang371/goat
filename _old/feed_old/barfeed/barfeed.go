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
	feed_old.BaseFeed
	frequencies       []common_old.Frequency
	useAdjustedValue  bool
	sType             series.Type
	defaultInstrument string
	currentBars       common_old.Bars
	lastBars          map[string][]common_old.Bar
}

func NewBaseBarFeed(frequencies []common_old.Frequency, sType series.Type, maxLen int) *baseBarFeed {
	baseFeed := feed_old.NewBaseFeed(maxLen)
	res := &baseBarFeed{
		BaseFeed:         *baseFeed,
		frequencies:      frequencies,
		useAdjustedValue: false,
		sType:            sType,
		lastBars:         map[string][]common_old.Bar{},
	}
	res.Self = res
	return res
}

func (b *baseBarFeed) Reset() {
	b.currentBars = nil
	b.lastBars = map[string][]common_old.Bar{}
}

func (b *baseBarFeed) GetCurrentBars() common_old.Bars {
	return b.currentBars
}

func (b *baseBarFeed) GetLastBar(instrument string) []common_old.Bar {
	if v, ok := b.lastBars[instrument]; ok {
		return v
	}
	return nil
}

func (b *baseBarFeed) GetNextBars() (common_old.Bars, error) {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) GetNextValues() (*time.Time, common_old.Bars, []common_old.Frequency, error) {
	bars, err := b.Self.(common_old.BarFeed).GetNextBars()
	if bars != nil && err == nil {
		freqList := bars.GetFrequencies()
		dateTime := bars.GetDateTime()

		if len(freqList) == 0 || dateTime == nil {
			lg.Logger.Error("invalid frequency and/or dateTime", zap.Any("Frequencies", freqList), zap.Any("DateTime", dateTime))
			return nil, nil, []common_old.Frequency{}, fmt.Errorf("invalid frequency and/or dateTime")
		}

		if b.currentBars != nil && b.currentBars.GetDateTime().After(*dateTime) {
			return nil, nil, []common_old.Frequency{},
				fmt.Errorf("bar date times are not in order. Previous dateTime was %s and current dateTime is %s",
					b.currentBars.GetDateTime(), dateTime)
		}

		b.currentBars = bars
		for _, v := range bars.GetInstruments() {
			b.lastBars[v] = bars.GetBarList(v)
		}
		return dateTime, bars, freqList, nil
	}
	// it is okay to return nil with no error. it is idle() case
	return nil, nil, []common_old.Frequency{}, err
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

func (b *baseBarFeed) GetFrequencies() []common_old.Frequency {
	return b.frequencies
}

func (b *baseBarFeed) CreateDataSeries(key string, maxLen int) common_old.BarDataSeries {
	// TODO: implement me and confirm if this is correct usage of dataseries
	ret := dataseries.NewBarDataSeries(b.sType, b.BaseFeed.GetMaxLen())
	return ret
}

func (b *baseBarFeed) GetDefaultInstrument() string {
	return b.defaultInstrument
}

func (b *baseBarFeed) GetRegisteredInstruments() []string {
	return b.GetKeys()
}

func (b *baseBarFeed) RegisterInstrument(instrument string, freqList []common_old.Frequency) error {
	b.defaultInstrument = instrument
	for _, freq := range freqList {
		err := b.RegisterDataSeries(instrument, freq)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *baseBarFeed) GetDataSeries(instrument string, freq common_old.Frequency) *series.Series {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}
