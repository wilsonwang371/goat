package barfeed

import (
	"goalgotrade/common"
	lg "goalgotrade/logger"
	"sync"
	"time"

	"github.com/go-gota/gota/series"
)

// TODO: make a LiveBarFeed interface and put it into common

type LiveBarFeed struct {
	baseBarFeed
	mu         sync.Mutex
	stopped    bool
	barsBuffer []common.Bars
	fetcher    common.LiveBarFetcher
}

func NewLiveBarFeed(freqList []common.Frequency, sType series.Type, maxLen int) *LiveBarFeed {
	res := LiveBarFeed{
		baseBarFeed: *NewBaseBarFeed(freqList, sType, maxLen),
	}
	res.Self = res
	return &res
}

func (l *LiveBarFeed) SetFetcher(f common.LiveBarFetcher) {
	if f != nil {
		l.fetcher = f
	} else {
		lg.Logger.Error("invalid fetcher was given")
	}
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
	return nil
}

func (l *LiveBarFeed) Stop() error {
	l.stopped = true
	return nil
}

func (l *LiveBarFeed) Join() error {
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
