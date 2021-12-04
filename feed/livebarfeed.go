package feed

import (
	"fmt"
	"goalgotrade/bar"
	"goalgotrade/consts/frequency"
	lg "goalgotrade/logger"
	"sync"
	"time"

	"github.com/go-gota/gota/series"

	"go.uber.org/zap"
)

// BarFetcher ...
type BarFetcher interface {
	RegisterInstrument(instrument string, freqList []frequency.Frequency) error
	GetInstrument() string
	GetFrequencies() []frequency.Frequency
	GetDSType() series.Type
	Start() error
	Stop() error
	PendingBarsC() <-chan bar.Bars
	IsRunning() bool
	ErrorC() <-chan error
	CurrentDateTime() *time.Time
}

// LiveBarFeed ...
type LiveBarFeed interface {
	BaseBarFeed
}

type liveBarFeed struct {
	baseBarFeed
	mu         sync.Mutex
	stopped    bool
	stopC      chan struct{}
	doneC      chan struct{}
	barsBuffer []bar.Bars
	fetcher    BarFetcher
}

// NewLiveBarFeed ...
func NewLiveBarFeed(f BarFetcher, maxLen int) LiveBarFeed {
	return newLiveBarFeed(f, maxLen)
}

func newLiveBarFeed(f BarFetcher, maxLen int) *liveBarFeed {
	if f == nil || len(f.GetInstrument()) == 0 || len(f.GetInstrument()) == 0 {
		lg.Logger.Error("invalid fetcher was given")
		return nil
	}
	res := &liveBarFeed{
		baseBarFeed: *newBaseBarFeed(f.GetFrequencies(), f.GetDSType(), maxLen),
		stopC:       make(chan struct{}, 1),
		doneC:       make(chan struct{}, 1),
		fetcher:     f,
	}
	return res
}

// IsLive ...
func (l *liveBarFeed) IsLive() bool {
	return true
}

// BarsHaveAdjClose ...
func (l *liveBarFeed) BarsHaveAdjClose(f BaseBarFeed) bool {
	return false
}

// GetCurrentDateTime ...
func (l *liveBarFeed) CurrentTime() *time.Time {
	// TODO: implement me
	return nil
}

// NextBars ...
func (l *liveBarFeed) NextBars() (bar.Bars, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.barsBuffer) != 0 {
		res := l.barsBuffer[0]
		l.barsBuffer = l.barsBuffer[1:]
		return res, nil
	}
	return nil, nil
}

// PeekDateTime ...
func (l *liveBarFeed) PeekDateTime() *time.Time {
	return nil
}

// Start ...
func (l *liveBarFeed) Start() error {
	if l.fetcher == nil {
		return fmt.Errorf("fetcher not set yet")
	}
	go func() {
		for {
			select {
			case bars := <-l.fetcher.PendingBarsC():
				if bars == nil {
					panic("invalid bars")
				}
				l.mu.Lock()
				l.barsBuffer = append(l.barsBuffer, bars)
				l.mu.Unlock()
			case <-l.stopC:
				l.stopped = true
				return
			}
		}
	}()
	return nil
}

// Stop ...
func (l *liveBarFeed) Stop() error {
	lg.Logger.Info("stopping live bar feed", zap.Any("LiveBarFeed", l))
	close(l.stopC)
	return nil
}

// Join ...
func (l *liveBarFeed) Join() error {
	<-l.doneC
	l.stopped = true
	return nil
}

// Eof ...
func (l *liveBarFeed) Eof() bool {
	return l.stopped == true
}
