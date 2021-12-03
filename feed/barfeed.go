package feed

import (
	"fmt"
	"goalgotrade/bar"
	"goalgotrade/consts/frequency"
	"goalgotrade/dataseries"
	lg "goalgotrade/logger"
	"time"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"
)

// BaseBarFeed ...
type BaseBarFeed interface {
	BaseFeed
	SetUseAdjustedValue(f BaseBarFeed, useAdjusted bool) error
	CurrentTime() *time.Time
	BarsHaveAdjClose(f BaseBarFeed) bool
	NextBars() (bar.Bars, error)
	AllFrequencies() []frequency.Frequency
	IsIntraDay() bool
	CurrentBars() bar.Bars
	LastBar(instrument string) bar.Bar
	DefaultInstrument() string
	RegisteredInstruments() []string
	RegisterInstrument(f BaseFeed, instrument string, freqList []frequency.Frequency) error
	DataSeries(instrument string, freq frequency.Frequency) dataseries.DataSeries
}

type baseBarFeed struct {
	baseFeed
	frequencies       []frequency.Frequency
	useAdjustedValue  bool
	sType             series.Type
	defaultInstrument string
	currentBars       bar.Bars
	lastBars          map[string]bar.Bar
}

// NewBaseBarFeed ...
func NewBaseBarFeed(frequencies []frequency.Frequency, sType series.Type, maxLen int) BaseBarFeed {
	return newBaseBarFeed(frequencies, sType, maxLen)
}

func newBaseBarFeed(frequencies []frequency.Frequency, sType series.Type, maxLen int) *baseBarFeed {
	p := newBaseFeed(maxLen)
	res := &baseBarFeed{
		baseFeed:         *p,
		frequencies:      frequencies,
		useAdjustedValue: false,
		sType:            sType,
		lastBars:         map[string]bar.Bar{},
	}
	return res
}

// Reset ...
func (b *baseBarFeed) Reset(f BaseFeed) error {
	b.currentBars = nil
	b.lastBars = map[string]bar.Bar{}
	return b.baseFeed.Reset(f)
}

// SetUseAdjustedValue ...
func (b *baseBarFeed) SetUseAdjustedValue(f BaseBarFeed, useAdjusted bool) error {
	if useAdjusted && !f.BarsHaveAdjClose(f) {
		return fmt.Errorf("not supported")
	}
	b.useAdjustedValue = useAdjusted
	for _, d := range b.baseFeed.registeredDs {
		bds := b.baseFeed.dataSeries[d.key][d.freq].(dataseries.BarDataSeries)
		bds.SetUseAdjustedValues(useAdjusted)
	}
	return nil
}

// CurrentTime ...
func (b *baseBarFeed) CurrentTime() *time.Time {
	panic("not implemented")
}

// BarsHaveAdjClose ...
func (b *baseBarFeed) BarsHaveAdjClose(f BaseBarFeed) bool {
	panic("not implemented")
}

// NextBars ...
func (b *baseBarFeed) NextBars() (bar.Bars, error) {
	panic("not implemented")
}

// CreateDataSeries ...
func (b *baseBarFeed) CreateDataSeries(key string, maxLen int) dataseries.DataSeries {
	// TODO: implement me and confirm if this is correct usage of data series
	ret := dataseries.NewBarDataSeries(b.sType, maxLen)
	ret.SetUseAdjustedValues(b.useAdjustedValue)
	return ret
}

// NextValues ...
func (b *baseBarFeed) NextValues(bf BaseFeed) (*time.Time, interface{}, []frequency.Frequency, error) {
	f := bf.(BaseBarFeed)
	bars, err := f.NextBars()
	if bars != nil && err == nil {
		freqList := bars.Frequencies()
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
		for _, v := range bars.Instruments() {
			b.lastBars[v] = bars.Bar(v)
		}
		return barsTime, bars, freqList, nil
	}
	// it is okay to return nil with no error. it is idle() case
	return nil, nil, []frequency.Frequency{}, err
}

// AllFrequencies ...
func (b *baseBarFeed) AllFrequencies() []frequency.Frequency {
	return b.frequencies
}

// IsIntraDay ...
func (b *baseBarFeed) IsIntraDay() bool {
	for _, v := range b.frequencies {
		if v < frequency.DAY {
			return true
		}
	}
	return false
}

// CurrentBars ...
func (b *baseBarFeed) CurrentBars() bar.Bars {
	return b.currentBars
}

// LastBar ...
func (b *baseBarFeed) LastBar(instrument string) bar.Bar {
	if v, ok := b.lastBars[instrument]; ok {
		return v
	}
	return nil
}

// DefaultInstrument ...
func (b *baseBarFeed) DefaultInstrument() string {
	return b.defaultInstrument
}

// RegisteredInstruments ...
func (b *baseBarFeed) RegisteredInstruments() []string {
	return b.Keys()
}

// RegisterInstrument ...
func (b *baseBarFeed) RegisterInstrument(f BaseFeed, instrument string, freqList []frequency.Frequency) error {
	b.defaultInstrument = instrument
	for _, freq := range freqList {
		err := f.RegisterDataSeries(f, instrument, freq)
		if err != nil {
			return err
		}
	}
	return nil
}

// DataSeries ...
func (b *baseBarFeed) DataSeries(instrument string, freq frequency.Frequency) dataseries.DataSeries {
	if instrument == "" {
		instrument = b.defaultInstrument
	}
	return b.baseFeed.dataSeries[instrument][freq]
}
