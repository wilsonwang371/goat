package dataseries

import (
	"goalgotrade/nugen/bar"
	"sync"
	"time"

	"github.com/go-gota/gota/series"
)

// BarDataSeries ...
type BarDataSeries interface {
	SequenceDataSeries
	OpenDS() series.Series
	HighDS() series.Series
	LowDS() series.Series
	CloseDS() series.Series
	AdjCloseDS() series.Series
	VolumeDS() series.Series
	PriceDS() series.Series
	ExtraDS() map[string]series.Series
	SetUseAdjustedValues(useAdjusted bool)
}

type barDataSeries struct {
	sequenceDataSeries
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

// NewBarDataSeries ...
func NewBarDataSeries(sType series.Type, maxLen int) BarDataSeries {
	res := &barDataSeries{
		sequenceDataSeries: *NewSequenceDataSeries(maxLen).(*sequenceDataSeries),
		open:               series.New(nil, sType, "open"),
		high:               series.New(nil, sType, "high"),
		low:                series.New(nil, sType, "low"),
		close:              series.New(nil, sType, "close"),
		adjClose:           series.New(nil, sType, "adjClose"),
		volume:             series.New(nil, series.Float, "volume"),
		extra:              map[string]series.Series{},
		useAdjValues:       false,
		maxLen:             maxLen,
		sType:              sType,
	}
	return res
}

func (s *barDataSeries) SetUseAdjustedValues(useAdjusted bool) {
	s.useAdjValues = useAdjusted
}

// Append ...
func (s *barDataSeries) Append(value interface{}) error {
	bar := value.(bar.Bar)
	return s.AppendWithDateTime(bar.Time(), bar)
}

// AppendWithDateTime ...
func (s *barDataSeries) AppendWithDateTime(dateTime *time.Time, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bar := value.(bar.Bar)

	err := bar.SetUseAdjustedValue(s.useAdjValues)
	if err != nil {
		return err
	}
	if err := s.sequenceDataSeries.AppendWithDateTime(dateTime, bar); err != nil {
		return err
	}

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

	// TODO: add extra columns
	return nil
}

// OpenDS ...
func (s *barDataSeries) OpenDS() series.Series {
	return s.open
}

// HighDS ...
func (s *barDataSeries) HighDS() series.Series {
	return s.high
}

// LowDS ...
func (s *barDataSeries) LowDS() series.Series {
	return s.low
}

// CloseDS ...
func (s *barDataSeries) CloseDS() series.Series {
	return s.close
}

// AdjCloseDS ...
func (s *barDataSeries) AdjCloseDS() series.Series {
	return s.adjClose
}

// VolumeDS ...
func (s *barDataSeries) VolumeDS() series.Series {
	return s.volume
}

// PriceDS ...
func (s *barDataSeries) PriceDS() series.Series {
	if s.useAdjValues {
		return s.adjClose
	}
	return s.close
}

// ExtraDS ...
func (s *barDataSeries) ExtraDS() map[string]series.Series {
	return s.extra
}

// Times ...
func (s *barDataSeries) Times() []*time.Time {
	return s.times
}
