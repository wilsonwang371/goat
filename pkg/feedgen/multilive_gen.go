package feedgen

import (
	"context"
	"reflect"
	"sync"
	"time"

	"goat/pkg/consts"
	"goat/pkg/core"
	"goat/pkg/logger"

	lg "goat/pkg/logger"

	"go.uber.org/zap"
)

const noDataSleepDuration = 100 * time.Millisecond

type MultiLiveBarFeedGenerator struct {
	ctx        context.Context
	bfg        core.FeedGenerator
	providers  []BarDataProvider
	pvdrChan   []chan core.Bars
	instrument string
	freq       []core.Frequency
	stopped    bool
}

// AppendNewValueToBuffer implements core.FeedGenerator
func (l *MultiLiveBarFeedGenerator) AppendNewValueToBuffer(t time.Time, v map[string]interface{}, f core.Frequency) error {
	// logger.Logger.Debug("LiveBarFeedGenerator::AppendNewValueToBuffer", zap.Any("t", t), zap.Any("v", v), zap.Any("f", f))
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
func (l *MultiLiveBarFeedGenerator) CreateDataSeries(key string, maxLen int) core.DataSeries {
	return l.bfg.CreateDataSeries(key, maxLen)
}

// Finish implements core.FeedGenerator
func (l *MultiLiveBarFeedGenerator) Finish() {
	l.bfg.Finish()
}

func (l *MultiLiveBarFeedGenerator) IsComplete() bool {
	return l.bfg.IsComplete()
}

// PeekNextTime implements core.FeedGenerator
func (l *MultiLiveBarFeedGenerator) PeekNextTime() *time.Time {
	return l.bfg.PeekNextTime()
}

// PopNextValues implements core.FeedGenerator
func (l *MultiLiveBarFeedGenerator) PopNextValues() (time.Time, map[string]interface{}, core.Frequency, error) {
	return l.bfg.PopNextValues()
}

func NewMultiLiveBarFeedGenerator(ctx context.Context, providers []BarDataProvider, instrument string,
	freq []core.Frequency,
	maxLen int,
) *MultiLiveBarFeedGenerator {
	if len(providers) == 0 {
		panic("providers is empty")
	}
	res := &MultiLiveBarFeedGenerator{
		ctx:        ctx,
		bfg:        core.NewBarFeedGenerator(freq, maxLen),
		providers:  providers,
		pvdrChan:   make([]chan core.Bars, len(providers)),
		instrument: instrument,
		freq:       freq,
		stopped:    false,
	}
	for i := range res.pvdrChan {
		res.pvdrChan[i] = make(chan core.Bars, 100)
	}
	return res
}

// start from here, we implement multiLiveBarFeedGenerator specific functions

func (l *MultiLiveBarFeedGenerator) SetInstrument(instrument string) {
	l.instrument = instrument
}

func (l *MultiLiveBarFeedGenerator) WaitAndRun(wg *sync.WaitGroup) error {
	wg.Wait()
	if err := l.Run(); err != nil {
		panic(err)
	}
	return nil
}

func (l *MultiLiveBarFeedGenerator) ProviderFetcher(idx int) {
	pvdr := l.providers[idx]
	pvdrChan := l.pvdrChan[idx]
	errorCount := 0
	for {
		select {
		case <-l.ctx.Done():
			return
		default:
		}
		if l.stopped {
			break
		}
		if bars, err := pvdr.nextBars(); err != nil {
			lg.Logger.Warn("nextBars failed", zap.Error(err),
				zap.Any("pvdr", pvdr),
				zap.Int("errorCount", errorCount))
			time.Sleep(consts.LiveGenFailureSleepDuration)
			errorCount++
		} else {
			if bars == nil {
				lg.Logger.Warn("got empty bars")
				continue
			}
			select {
			case pvdrChan <- bars:
				errorCount = 0
			default:
				lg.Logger.Warn("pvdrChan is full, drop bars", zap.Any("bars", bars))
			}
		}
	}
}

func singleBarFromBars(bars core.Bars) core.Bar {
	if bars == nil {
		panic("bars is nil")
	}
	for _, bar := range bars {
		return bar
	}
	panic("bars is empty")
}

func (l *MultiLiveBarFeedGenerator) Run() error {
	for i, p := range l.providers {
		logger.Logger.Debug("start provider", zap.Any("p", reflect.TypeOf(p)))
		if err := p.init(l.instrument, l.freq); err != nil {
			logger.Logger.Error("failed to init provider", zap.Error(err))
			return err
		}
		if err := p.connect(); err != nil {
			logger.Logger.Error("failed to connect provider", zap.Error(err))
			return err
		}
		go l.ProviderFetcher(i)
	}

	pendingBars := make([]core.Bars, len(l.providers))

	for {
		select {
		case <-l.ctx.Done():
			return nil
		default:
		}
		if l.stopped {
			logger.Logger.Info("stopped")
			break
		}

		for i, pvdrChan := range l.pvdrChan {
			if pendingBars[i] == nil {
				select {
				case pendingBars[i] = <-pvdrChan:
				default:
				}
			}
		}

		// find the earliest bar
		var earliestBarIdx int = -1
		for i, b := range pendingBars {
			if b == nil {
				continue
			}

			if earliestBarIdx == -1 {
				earliestBarIdx = i
			} else if singleBarFromBars(b).DateTime().Before(
				singleBarFromBars(pendingBars[earliestBarIdx]).DateTime()) {
				earliestBarIdx = i
			}
		}

		if earliestBarIdx == -1 {
			// no bar available
			time.Sleep(noDataSleepDuration)
			continue
		} else {
			// we have a bar, append it to buffer
			bars := pendingBars[earliestBarIdx]
			// logger.Logger.Debug("got bar", zap.Any("bar", bars))
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
			pendingBars[earliestBarIdx] = nil
		}
	}
	return nil
}

func (l *MultiLiveBarFeedGenerator) Stop() error {
	for _, p := range l.providers {
		if err := p.stop(); err != nil {
			return err
		}
	}
	l.stopped = true
	return nil
}
