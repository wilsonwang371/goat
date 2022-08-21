package feedgen

import (
	"sync"
	"time"

	"goat/pkg/consts"
	"goat/pkg/core"
	"goat/pkg/logger"

	lg "goat/pkg/logger"

	"github.com/go-gota/gota/series"
	"go.uber.org/zap"
)

type BarDataProvider interface {
	init(instrument string, freqList []core.Frequency) error
	connect() error
	nextBars() (core.Bars, error) // this can return nothing but with no error, you should not block this forever
	reset() error
	stop() error
	datatype() series.Type
}

type LiveBarFeedGenerator struct {
	bfg        core.FeedGenerator
	provider   BarDataProvider
	instrument string
	freq       []core.Frequency
	stopped    bool
}

// AppendNewValueToBuffer implements core.FeedGenerator
func (l *LiveBarFeedGenerator) AppendNewValueToBuffer(t time.Time, v map[string]interface{}, f core.Frequency) error {
	logger.Logger.Debug("LiveBarFeedGenerator::AppendNewValueToBuffer", zap.Any("t", t), zap.Any("v", v), zap.Any("f", f))
	for {
		if err := l.bfg.AppendNewValueToBuffer(t, v, f); err != nil {
			logger.Logger.Debug("LiveBarFeedGenerator::AppendNewValueToBuffer failed", zap.Error(err))
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return nil
}

// CreateDataSeries implements core.FeedGenerator
func (l *LiveBarFeedGenerator) CreateDataSeries(key string, maxLen int) core.DataSeries {
	return l.bfg.CreateDataSeries(key, maxLen)
}

// Finish implements core.FeedGenerator
func (l *LiveBarFeedGenerator) Finish() {
	l.bfg.Finish()
}

func (l *LiveBarFeedGenerator) IsComplete() bool {
	return l.bfg.IsComplete()
}

// PeekNextTime implements core.FeedGenerator
func (l *LiveBarFeedGenerator) PeekNextTime() *time.Time {
	return l.bfg.PeekNextTime()
}

// PopNextValues implements core.FeedGenerator
func (l *LiveBarFeedGenerator) PopNextValues() (time.Time, map[string]interface{}, core.Frequency, error) {
	return l.bfg.PopNextValues()
}

func NewLiveBarFeedGenerator(provider BarDataProvider, instrument string,
	freq []core.Frequency,
	maxLen int,
) *LiveBarFeedGenerator {
	res := &LiveBarFeedGenerator{
		bfg:        core.NewBarFeedGenerator(freq, maxLen),
		provider:   provider,
		instrument: instrument,
		freq:       freq,
		stopped:    false,
	}
	return res
}

// start from here, we implement liveBarFeedGenerator specific functions

func (l *LiveBarFeedGenerator) SetInstrument(instrument string) {
	l.instrument = instrument
}

func (l *LiveBarFeedGenerator) WaitAndRun(wg *sync.WaitGroup) error {
	wg.Wait()
	l.Run()
	return nil
}

func (l *LiveBarFeedGenerator) Run() error {
	if l.provider == nil {
		panic("provider is nil")
	}

	if err := l.provider.init(l.instrument, l.freq); err != nil {
		logger.Logger.Error("failed to init provider", zap.Error(err))
		return err
	}
	if err := l.provider.connect(); err != nil {
		logger.Logger.Error("failed to connect provider", zap.Error(err))
		return err
	}

	errorCount := 0

	for {
		if l.stopped {
			break
		}
		if bars, err := l.provider.nextBars(); err != nil {
			lg.Logger.Error("nextBars failed", zap.Error(err))
			time.Sleep(consts.LiveGenFailureSleepDuration)
			errorCount++
			if errorCount > consts.LiveGenFailureMaxCount {
				lg.Logger.Error("too many errors, stop")
				return err
			}
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

			// reset error count
			errorCount = 0
		}
	}
	return nil
}

func (l *LiveBarFeedGenerator) Stop() error {
	if err := l.provider.stop(); err != nil {
		return err
	}
	l.stopped = true
	return nil
}
