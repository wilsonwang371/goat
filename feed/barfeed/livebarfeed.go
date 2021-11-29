package barfeed

import (
	"fmt"
	"goalgotrade/common"
	lg "goalgotrade/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TODO: make a LiveBarFeed interface and put it into common

type LiveBarFeed struct {
	baseBarFeed
	mu         sync.Mutex
	stopped    bool
	stopC      chan struct{}
	doneC      chan struct{}
	barsBuffer []common.Bars
	fetcher    common.LiveBarFetcher
}

func NewLiveBarFeed(f common.LiveBarFetcher, maxLen int) *LiveBarFeed {
	if f == nil || len(f.GetInstrument()) == 0 || len(f.GetInstrument()) == 0 {
		lg.Logger.Error("invalid fetcher was given")
		return nil
	}
	res := &LiveBarFeed{
		baseBarFeed: *NewBaseBarFeed(f.GetFrequencies(), f.GetDSType(), maxLen),
		stopC:       make(chan struct{}, 1),
		doneC:       make(chan struct{}, 1),
		fetcher:     f,
	}
	res.Self = res
	return res
}

func (l *LiveBarFeed) IsLive() bool {
	return true
}

func (l *LiveBarFeed) BarsHaveAdjClose() bool {
	return false
}

func (l *LiveBarFeed) GetCurrentDateTime() *time.Time {
	// TODO: implement me
	return nil
}

func (l *LiveBarFeed) GetNextBars() (common.Bars, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.barsBuffer) != 0 {
		res := l.barsBuffer[0]
		l.barsBuffer = l.barsBuffer[1:]
		return res, nil
	}
	return nil, nil
}

func (l *LiveBarFeed) PeekDateTime() *time.Time {
	return nil
}

func (l *LiveBarFeed) Start() error {
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

func (l *LiveBarFeed) Stop() error {
	lg.Logger.Info("stopping live bar feed", zap.Any("LiveBarFeed", l))
	close(l.stopC)
	return nil
}

func (l *LiveBarFeed) Join() error {
	<-l.doneC
	l.stopped = true
	return nil
}

func (l *LiveBarFeed) Eof() bool {
	return l.stopped == true
}

func (l *LiveBarFeed) Dispatch() (bool, error) {
	var res bool

	// TODO: implement me

	res, err := l.baseBarFeed.Dispatch()
	return res, err
}
