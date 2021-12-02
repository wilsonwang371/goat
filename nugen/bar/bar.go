package bar

import (
	"fmt"
	"goalgotrade/nugen/consts/frequency"
	"time"

	lg "goalgotrade/nugen/logger"

	"go.uber.org/zap"
)

// Bar ...
type Bar interface {
	Open() float64
	High() float64
	Low() float64
	Close() float64
	Volume() float64
	AdjClose() float64
	Frequency() frequency.Frequency
	Price() float64
	UseAdjValue() bool
	SetUseAdjustedValue(useAdjusted bool) error
	Time() *time.Time
}

type basicBar struct {
	barTime          *time.Time
	open             float64
	high             float64
	low              float64
	close            float64
	adjClose         float64
	volume           float64
	frequency        frequency.Frequency
	useAdjustedValue bool
}

// NewBasicBar ...
func NewBasicBar(barTime time.Time, o, h, l, c, v, adjClose float64, freq frequency.Frequency) (Bar, error) {
	if h < l {
		lg.Logger.Error("high < low on %s", zap.Time("barTime", barTime))
		return nil, fmt.Errorf("high < low ")
	} else if h < o {
		lg.Logger.Error("high < open on %s", zap.Time("barTime", barTime))
		return nil, fmt.Errorf("high < open")
	} else if h < c {
		lg.Logger.Error("high < close on %s", zap.Time("barTime", barTime))
		return nil, fmt.Errorf("high < close")
	} else if l > o {
		lg.Logger.Error("low > open on %s", zap.Time("barTime", barTime))
		return nil, fmt.Errorf("low > open")
	} else if l > c {
		lg.Logger.Error("low > close on %s", zap.Time("barTime", barTime))
		return nil, fmt.Errorf("low > close")
	}
	tmpTime := barTime
	return &basicBar{
		barTime:          &tmpTime,
		open:             o,
		high:             h,
		low:              l,
		close:            c,
		adjClose:         adjClose,
		volume:           v,
		frequency:        freq,
		useAdjustedValue: false,
	}, nil
}

// SetUseAdjustedValue ...
func (b *basicBar) SetUseAdjustedValue(useAdjusted bool) error {
	b.useAdjustedValue = useAdjusted
	return nil
}

// UseAdjValue ...
func (b *basicBar) UseAdjValue() bool {
	return b.useAdjustedValue
}

// Time ...
func (b *basicBar) Time() *time.Time {
	return b.barTime
}

// Open ...
func (b *basicBar) Open() float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.open / b.close
	}
	return b.open
}

// High ...
func (b *basicBar) High() float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.high / b.close
	}
	return b.high
}

// Low ...
func (b *basicBar) Low() float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.low / b.close
	}
	return b.low
}

// Close ...
func (b *basicBar) Close() float64 {
	if b.useAdjustedValue {
		return b.adjClose
	}
	return b.open
}

// Volume ...
func (b *basicBar) Volume() float64 {
	return b.volume
}

// AdjClose ...
func (b *basicBar) AdjClose() float64 {
	return b.adjClose
}

// Frequency ...
func (b *basicBar) Frequency() frequency.Frequency {
	return b.frequency
}

// Price ...
func (b *basicBar) Price() float64 {
	if b.useAdjustedValue {
		return b.adjClose
	}
	return b.open
}

// Bars ...
type Bars interface {
	Instruments() []string
	Frequencies() []frequency.Frequency
	Time() *time.Time
	Bar(instrument string) Bar
	Items(instrument string) []Bar
	AddBarList(instrument string, barList []Bar) error
}

type bars struct {
	barList map[string]Bar
	barTime *time.Time
}

// NewBars ...
func NewBars() Bars {
	return &bars{
		barList: map[string]Bar{},
	}
}

// Instruments ...
func (b *bars) Instruments() []string {
	keys := make([]string, len(b.barList))
	i := 0
	for k := range b.barList {
		keys[i] = k
		i++
	}
	return keys
}

// Frequencies ...
func (b *bars) Frequencies() []frequency.Frequency {
	freqs := map[frequency.Frequency]struct{}{}
	for _, bar := range b.barList {
		freqs[bar.Frequency()] = struct{}{}
	}
	res := []frequency.Frequency{}
	for f := range freqs {
		res = append(res, f)
	}
	return res
}

// Time ...
func (b *bars) Time() *time.Time {
	return b.barTime
}

// Bar ...
func (b *bars) Bar(instrument string) Bar {
	if val, ok := b.barList[instrument]; ok {
		return val
	}
	return nil
}

// Items ...
func (b *bars) Items(instrument string) []Bar {
	res := []Bar{}
	if bar, ok := b.barList[instrument]; ok {
		res = append(res, bar)
	}
	return res
}

func (b *bars) addSingleBar(instrument string, bar Bar) error {
	if _, ok := b.barList[instrument]; ok {
		lg.Logger.Error("instrument exists already", zap.String("instrument", instrument))
		return fmt.Errorf("instrument exists already %s", instrument)
	}
	b.barList[instrument] = bar

	if b.barTime == nil || bar.Time().Before(*b.barTime) {
		b.barTime = bar.Time()
	}
	return nil
}

// AddBarList ...
func (b *bars) AddBarList(instrument string, barList []Bar) error {
	for _, v := range barList {
		err := b.addSingleBar(instrument, v)
		if err != nil {
			return err
		}
	}
	return nil
}
