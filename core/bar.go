package core

import (
	"fmt"
	"goalgotrade/common"
	"time"

	lg "goalgotrade/logger"

	"go.uber.org/zap"
)

type bars struct {
	bars        map[string]common.Bar
	datetime    *time.Time
	frequencies []common.Frequency
}

func NewBars() common.Bars {
	return &bars{}
}

func (b *bars) GetInstruments() []string {
	keys := make([]string, len(b.bars))
	i := 0
	for k := range b.bars {
		keys[i] = k
		i++
	}
	return keys
}

func (b *bars) GetFrequencies() []common.Frequency {
	return b.frequencies
}

func (b *bars) GetBar(instrument string) common.Bar {
	if bar, ok := b.bars[instrument]; ok {
		return bar
	}
	return nil
}

func (b *bars) GetDateTime() *time.Time {
	return b.datetime
}

func (b *bars) AddBar(instrument string, bar common.Bar) error {
	addFrequency := true
	if _, ok := b.bars[instrument]; !ok {
		for _, v := range b.frequencies {
			if v == bar.Frequency() {
				addFrequency = false
			}
		}
		if addFrequency {
			b.frequencies = append(b.frequencies, bar.Frequency())
		}
		b.bars[instrument] = bar
		return nil
	}
	return fmt.Errorf("bar exists already")
}

type basicBar struct {
	datetime         *time.Time
	open             float64
	high             float64
	low              float64
	close            float64
	adjClose         float64
	volume           int
	frequency        common.Frequency
	useAdjustedValue bool
}

func NewBasicBar(datetime time.Time, o, h, l, c float64, v int, adjClose float64, freq common.Frequency) common.Bar {
	if h < l {
		lg.Logger.Error("high < low on %s", zap.Time("datetime", datetime))
		return nil
	} else if h < o {
		lg.Logger.Error("high < open on %s", zap.Time("datetime", datetime))
		return nil
	} else if h < c {
		lg.Logger.Error("high < close on %s", zap.Time("datetime", datetime))
		return nil
	} else if l > o {
		lg.Logger.Error("low > open on %s", zap.Time("datetime", datetime))
		return nil
	} else if l > c {
		lg.Logger.Error("low > close on %s", zap.Time("datetime", datetime))
		return nil
	}
	tmptime := datetime
	return &basicBar{
		datetime:         &tmptime,
		open:             o,
		high:             h,
		low:              l,
		close:            c,
		adjClose:         adjClose,
		volume:           v,
		frequency:        freq,
		useAdjustedValue: false,
	}
}

func (b *basicBar) SetUseAdjustedValue(useAdjusted bool) error {
	b.useAdjustedValue = useAdjusted
	return nil
}

func (b *basicBar) GetUseAdjValue() bool {
	return b.useAdjustedValue
}

func (b *basicBar) GetDateTime() *time.Time {
	return b.datetime
}

func (b *basicBar) Open(adjusted bool) float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.open / b.close
	}
	return b.open
}

func (b *basicBar) High(adjusted bool) float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.high / b.close
	}
	return b.high
}

func (b *basicBar) Low(adjusted bool) float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.low / b.close
	}
	return b.low
}

func (b *basicBar) Close(adjusted bool) float64 {
	if adjusted {
		return b.adjClose
	}
	return b.open
}

func (b *basicBar) Volume() int {
	return b.volume
}

func (b *basicBar) AdjClose() float64 {
	return b.adjClose
}

func (b *basicBar) Frequency() common.Frequency {
	return b.frequency
}

func (b *basicBar) Price() float64 {
	if b.useAdjustedValue {
		return b.adjClose
	}
	return b.open
}
