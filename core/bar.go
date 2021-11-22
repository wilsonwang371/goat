package core

import (
	"goalgotrade/common"
	"time"

	lg "goalgotrade/logger"

	"go.uber.org/zap"
)

type bars struct {
	barList     map[string][]common.Bar
	dateTime    *time.Time
	frequencies []common.Frequency
}

func NewBars() common.Bars {
	return &bars{
		barList: make(map[string][]common.Bar),
	}
}

func (b *bars) GetInstruments() []string {
	keys := make([]string, len(b.barList))
	i := 0
	for k := range b.barList {
		keys[i] = k
		i++
	}
	return keys
}

func (b *bars) GetFrequencies() []common.Frequency {
	return b.frequencies
}

func (b *bars) GetBarList(instrument string) []common.Bar {
	if bar, ok := b.barList[instrument]; ok {
		return bar
	}
	return nil
}

func (b *bars) GetDateTime() *time.Time {
	return b.dateTime
}

func (b *bars) addSingleBar(instrument string, bar common.Bar) error {
	addFrequency := true
	if _, ok := b.barList[instrument]; ok {
		lg.Logger.Warn("instrument exists already", zap.String("instrument", instrument))
	}
	for _, v := range b.frequencies {
		if v == bar.Frequency() {
			addFrequency = false
		}
	}
	if addFrequency {
		b.frequencies = append(b.frequencies, bar.Frequency())
	}
	b.barList[instrument] = append(b.barList[instrument], bar)
	return nil
}

func (b *bars) AddBarList(instrument string, barList []common.Bar) error {
	for _, v := range barList {
		err := b.addSingleBar(instrument, v)
		if err != nil {
			return err
		}
	}
	return nil
}

type basicBar struct {
	dateTime         *time.Time
	open             float64
	high             float64
	low              float64
	close            float64
	adjClose         float64
	volume           float64
	frequency        common.Frequency
	useAdjustedValue bool
}

func NewBasicBar(dateTime time.Time, o, h, l, c, v, adjClose float64, freq common.Frequency) common.Bar {
	if h < l {
		lg.Logger.Error("high < low on %s", zap.Time("datetime", dateTime))
		return nil
	} else if h < o {
		lg.Logger.Error("high < open on %s", zap.Time("datetime", dateTime))
		return nil
	} else if h < c {
		lg.Logger.Error("high < close on %s", zap.Time("datetime", dateTime))
		return nil
	} else if l > o {
		lg.Logger.Error("low > open on %s", zap.Time("datetime", dateTime))
		return nil
	} else if l > c {
		lg.Logger.Error("low > close on %s", zap.Time("datetime", dateTime))
		return nil
	}
	tmptime := dateTime
	return &basicBar{
		dateTime:         &tmptime,
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
	return b.dateTime
}

func (b *basicBar) Open() float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.open / b.close
	}
	return b.open
}

func (b *basicBar) High() float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.high / b.close
	}
	return b.high
}

func (b *basicBar) Low() float64 {
	if b.useAdjustedValue {
		return b.adjClose * b.low / b.close
	}
	return b.low
}

func (b *basicBar) Close() float64 {
	if b.useAdjustedValue {
		return b.adjClose
	}
	return b.open
}

func (b *basicBar) Volume() float64 {
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
