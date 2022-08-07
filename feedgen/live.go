package feedgen

import (
	"goalgotrade/core"
	lg "goalgotrade/logger"
	"time"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"
)

type BarFetcherProvider interface {
	init(instrument string, freqList []core.Frequency) error
	connect() error
	nextBars() (map[string]core.Bar, error)
	reset() error
	stop() error
	datatype() series.Type
}

type LiveBarFeedGenerator struct {
	bfg        core.FeedGenerator
	provider   BarFetcherProvider
	instrument string
	freq       []core.Frequency
}

// AppendNewValueToBuffer implements core.FeedGenerator
func (l *LiveBarFeedGenerator) AppendNewValueToBuffer(t time.Time, v map[string]interface{}, f core.Frequency) error {
	return l.bfg.AppendNewValueToBuffer(t, v, f)
}

// CreateDataSeries implements core.FeedGenerator
func (l *LiveBarFeedGenerator) CreateDataSeries(key string, maxLen int) core.DataSeries {
	return l.bfg.CreateDataSeries(key, maxLen)
}

// Finish implements core.FeedGenerator
func (l *LiveBarFeedGenerator) Finish() {
	l.bfg.Finish()
}

// PeekNextTime implements core.FeedGenerator
func (l *LiveBarFeedGenerator) PeekNextTime() *time.Time {
	return l.bfg.PeekNextTime()
}

// PopNextValues implements core.FeedGenerator
func (l *LiveBarFeedGenerator) PopNextValues() (time.Time, map[string]interface{}, core.Frequency, error) {
	return l.bfg.PopNextValues()
}

func NewLiveBarFeedGenerator(provider BarFetcherProvider, instrument string, freq []core.Frequency, maxLen int) core.FeedGenerator {
	return &LiveBarFeedGenerator{
		bfg:        core.NewBarFeedGenerator(freq, maxLen),
		provider:   provider,
		instrument: instrument,
		freq:       freq,
	}
}

// start from here, we implement liveBarFeedGenerator specific functions

func (l *LiveBarFeedGenerator) Run() error {
	if err := l.provider.init(l.instrument, l.freq); err != nil {
		return err
	}

	for {
		if bars, err := l.provider.nextBars(); err != nil {
			lg.Logger.Error("nextBars failed", zap.Error(err))
			return err
		} else {
			if bars == nil {
				lg.Logger.Warn("got empty bars")
				continue
			}
			var freq *core.Frequency
			for _, v := range bars {
				if freq == nil {
					f := v.(core.Bar).Frequency()
					freq = &f
				}
				if *freq != v.(core.Bar).Frequency() {
					panic("freq mismatch")
				}
			}
			res := make(map[string]interface{}, len(bars))
			for k, v := range bars {
				res[k] = v
			}
			l.AppendNewValueToBuffer(time.Time{}, res, *freq)
		}
	}
}

func (l *LiveBarFeedGenerator) Stop() error {
	if err := l.provider.stop(); err != nil {
		return err
	}
	return nil
}
