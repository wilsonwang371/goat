package fetcher

import (
	"fmt"
	"goalgotrade/common"
	"sync"
	"time"
)

const DefaultDataPullInterval = 10 * time.Second

type NetworkBarFetcher struct {
	mu           sync.Mutex
	instrument   string
	freqList     []common.Frequency
	pendingBars  chan common.Bars
	stopped      bool
	stopC        chan struct{}
	doneC        chan struct{}
	pullInterval time.Duration
}

// TODO: implement me

func NewNetworkBarFetcher(pullInterval time.Duration) common.LiveBarFetcher {
	n := &NetworkBarFetcher{
		instrument:   "",
		pendingBars:  make(chan common.Bars, 32),
		stopped:      true,
		stopC:        make(chan struct{}, 1),
		doneC:        make(chan struct{}, 1),
		pullInterval: DefaultDataPullInterval,
	}
	if pullInterval != 0 {
		n.pullInterval = pullInterval
	}
	return n
}

func (n *NetworkBarFetcher) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.stopped {
		return fmt.Errorf("already running")
	}
	n.stopped = false
	go n.run()
	return nil
}

func (n *NetworkBarFetcher) run() {
	defer func() {
		close(n.doneC)
		n.stopped = true
	}()
	t := time.NewTimer(n.pullInterval)
	for {
		select {
		case <-t.C:
		case <-n.stopC:
			break
		}
		// TODO: implement me
		t.Reset(n.pullInterval)
	}
}

func (n *NetworkBarFetcher) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.stopped {
		return fmt.Errorf("already stopped")
	}
	close(n.stopC)
	<-n.doneC
	return nil
}

func (n *NetworkBarFetcher) RegisterInstrument(instrument string, freqList []common.Frequency) error {
	if !n.stopped {
		return fmt.Errorf("fetcher is already running")
	}
	if len(freqList) == 0 {
		return fmt.Errorf("empty list of frequencies")
	}
	if n.instrument == "" {
		n.instrument = instrument
		n.freqList = freqList
		return nil
	}
	return fmt.Errorf("cannot only register instrument once")
}

func (n *NetworkBarFetcher) CurrentDateTime() *time.Time {
	return nil
}

func (n *NetworkBarFetcher) ErrorC() <-chan error {
	panic("implement me")
}

func (n *NetworkBarFetcher) PendingBarsC() <-chan common.Bars {
	return n.pendingBars
}

func (n *NetworkBarFetcher) IsRunning() bool {
	return !n.stopped
}
