package dataseries

import (
	"goalgotrade/common"
	"sync"
	"time"

	"github.com/go-gota/gota/series"
)

type barDataSeries struct {
	mu           sync.RWMutex
	open         series.Series
	high         series.Series
	low          series.Series
	close        series.Series
	adjClose     series.Series
	volume       series.Series
	extra        map[string]series.Series
	useAdjValues bool
	maxLen       int
	sType        series.Type
}

func NewBarDataSeries(sType series.Type, maxLen int) common.BarDataSeries {
	return &barDataSeries{
		open:         series.New(nil, sType, "open"),
		high:         series.New(nil, sType, "high"),
		low:          series.New(nil, sType, "low"),
		close:        series.New(nil, sType, "close"),
		adjClose:     series.New(nil, sType, "adjClose"),
		volume:       series.New(nil, series.Float, "volume"),
		extra:        map[string]series.Series{},
		useAdjValues: false,
		maxLen:       maxLen,
		sType:        sType,
	}
}

func (s *barDataSeries) Append(bar common.Bar) error {
	return s.AppendWithDateTime(*bar.GetDateTime(), bar)
}

func (s *barDataSeries) AppendWithDateTime(dateTime time.Time, bar common.Bar) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := bar.SetUseAdjustedValue(s.useAdjValues)
	if err != nil {
		return err
	}
	// TODO: call super class?

	s.open.Append(bar.Open())
	s.high.Append(bar.High())
	s.low.Append(bar.Low())
	s.close.Append(bar.Close())
	s.adjClose.Append(bar.AdjClose())
	s.volume.Append(bar.Volume())

	for _, val := range []*series.Series{&s.open, &s.high, &s.low, &s.close, &s.adjClose, &s.volume} {
		if val.Len() > s.maxLen {
			*val = val.Slice(val.Len()-s.maxLen, s.maxLen-1)
		}
	}

	newExtra := map[string]series.Series{}
	for k, v := range s.extra {
		if v.Len() > s.maxLen {
			newExtra[k] = v.Slice(v.Len()-s.maxLen, s.maxLen-1)
		} else {
			newExtra[k] = v
		}
	}
	s.extra = newExtra

	// TODO: add extr columns
	return nil
}

func (s *barDataSeries) OpenDS() *series.Series {
	return &s.open
}

func (s *barDataSeries) HighDS() *series.Series {
	return &s.high
}

func (s *barDataSeries) LowDS() *series.Series {
	return &s.low
}

func (s *barDataSeries) CloseDS() *series.Series {
	return &s.close
}

func (s *barDataSeries) AdjCloseDS() *series.Series {
	return &s.adjClose
}

func (s *barDataSeries) VolumeDS() *series.Series {
	return &s.volume
}

func (s *barDataSeries) PriceDS() *series.Series {
	if s.useAdjValues {
		return &s.adjClose
	}
	return &s.close
}

func (s *barDataSeries) ExtraDS() map[string]series.Series {
	return s.extra
}
