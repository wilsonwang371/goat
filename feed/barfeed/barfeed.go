package barfeed

import (
	"goalgotrade/common"
	"goalgotrade/dataseries"
	"goalgotrade/feed"
	lg "goalgotrade/logger"
	"time"

	"github.com/go-gota/gota/series"
)

type baseBarFeed struct {
	*feed.BaseFeed
	frequencies      []common.Frequency
	useAdjustedValue bool
	stype            series.Type
}

func NewBaseBarFeed(frequencies []common.Frequency, stype series.Type, maxlen int) common.BarFeed {
	return &baseBarFeed{
		BaseFeed:         feed.NewBaseFeed(maxlen),
		frequencies:      frequencies,
		useAdjustedValue: false,
		stype:            stype,
	}
}

func (b *baseBarFeed) GetCurrentBars() common.Bars {
	// TODO: Implement me
	return nil
}

func (b *baseBarFeed) GetLastBar() common.Bar {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) GetNextBars() common.Bars {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) GetNextValues() (*time.Time, common.Bars, common.Frequency, error) {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
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
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) GetRegisteredInstruments() []string {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) RegisterInstrument(instrument string, freq common.Frequency) error {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}

func (b *baseBarFeed) GetDataSeries(instrument string, freq common.Frequency) *series.Series {
	// TODO: implement me
	lg.Logger.Error("not implemented")
	panic("not implemented")
}
