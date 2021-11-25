package fetcher

import (
	"goalgotrade/common"
	"time"
)

type WebFetcher struct {
	instruments  []string
	freqList     []common.Frequency
	pendingBars  chan common.Bars
	stopped      bool
	stopC        chan struct{}
	doneC        chan struct{}
	pullInterval time.Duration
}

const DefaultDataPullInterval = 10 * time.Second

// TODO: implement me

func NewWebFetcher(instruments []string, freqList []common.Frequency, pullInterval time.Duration) common.LiveBarFetcher {
	w := &WebFetcher{
		instruments:  instruments,
		freqList:     freqList,
		pendingBars:  make(chan common.Bars, 32),
		stopped:      false,
		stopC:        make(chan struct{}, 1),
		doneC:        make(chan struct{}, 1),
		pullInterval: DefaultDataPullInterval,
	}
	if pullInterval != 0 {
		w.pullInterval = pullInterval
	}
	go w.run()
	return w
}

func (w *WebFetcher) run() {
	defer func() {
		close(w.doneC)
		w.stopped = true
	}()
	t := time.NewTimer(w.pullInterval)
	for {
		select {
		case <-t.C:
		case <-w.stopC:
			break
		}
		// TODO: implement me
		t.Reset(w.pullInterval)
	}
}

func (w *WebFetcher) Stop() {
	close(w.stopC)
	<-w.doneC
}

func (w *WebFetcher) CurrentDateTime() *time.Time {
	return nil
}

func (w *WebFetcher) ErrorC() <-chan error {
	panic("implement me")
}

func (w *WebFetcher) PendingBarsC() <-chan common.Bars {
	return w.pendingBars
}

func (w *WebFetcher) IsRunning() bool {
	return !w.stopped
}
