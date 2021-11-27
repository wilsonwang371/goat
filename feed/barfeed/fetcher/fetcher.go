package fetcher

import (
	"fmt"
	"goalgotrade/common"
	lg "goalgotrade/logger"
	"sync"
	"time"
)

const DefaultDataPullInterval = 10 * time.Second

type BaseBarFetcher struct {
	Self         interface{}
	mu           sync.Mutex
	instrument   string
	freqList     []common.Frequency
	pendingBars  chan common.Bars
	stopped      bool
	stopC        chan struct{}
	doneC        chan struct{}
	errorC       chan error
	pullInterval time.Duration
}

type BarFetcherProvider interface {
	init(instrument string, freqList []common.Frequency) error
	connect() error
	nextBars() (common.Bars, error)
	reset() error
	stop() error
}

// TODO: implement me

func NewBaseBarFetcher(pullInterval time.Duration) *BaseBarFetcher {
	b := &BaseBarFetcher{
		instrument:   "",
		pendingBars:  make(chan common.Bars, 32),
		stopped:      true,
		stopC:        make(chan struct{}, 1),
		doneC:        make(chan struct{}, 1),
		errorC:       make(chan error, 8),
		pullInterval: DefaultDataPullInterval,
	}
	if pullInterval >= 0 {
		b.pullInterval = pullInterval
	}
	b.Self = b
	return b
}

func (b *BaseBarFetcher) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.stopped {
		return fmt.Errorf("already running")
	}
	if b.instrument == "" || len(b.freqList) == 0 {
		return fmt.Errorf("no instrument registered")
	}
	b.stopped = false
	if err := b.Self.(BarFetcherProvider).init(b.instrument, b.freqList); err != nil {
		b.stopped = true
		return err
	}
	go b.run()
	return nil
}

func (b *BaseBarFetcher) appendError(err error) error {
	select {
	case b.errorC <- err:
	default:
		return fmt.Errorf("failed to append error")
	}
	return nil
}

func (b *BaseBarFetcher) run() {
	defer func() {
		close(b.doneC)
		b.stopped = true
	}()
	t := time.NewTimer(b.pullInterval)
	if err := b.Self.(BarFetcherProvider).connect(); err != nil {
		b.appendError(err)
		return
	}
	for {
		select {
		case <-t.C:
			t.Reset(b.pullInterval)
		case <-b.stopC:
			return
		}

		if bars, err := b.Self.(BarFetcherProvider).nextBars(); err != nil {
			return
		} else {
			if bars == nil {
				lg.Logger.Warn("got empty bars")
				continue
			}
			select {
			case b.pendingBars <- bars:
			default:
				lg.Logger.Error("pendingBars are full")
				b.appendError(fmt.Errorf("pendingBars are full"))
			}
		}
	}
}

func (b *BaseBarFetcher) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stopped {
		return fmt.Errorf("already stopped")
	}
	if err := b.Self.(BarFetcherProvider).stop(); err != nil {
		return err
	}
	close(b.stopC)
	<-b.doneC
	return nil
}

func (b *BaseBarFetcher) RegisterInstrument(instrument string, freqList []common.Frequency) error {
	if !b.stopped {
		return fmt.Errorf("fetcher is already running")
	}
	if len(freqList) == 0 {
		return fmt.Errorf("empty list of frequencies")
	}
	if b.instrument == "" {
		b.instrument = instrument
		b.freqList = freqList
		return nil
	}
	return fmt.Errorf("cannot only register instrument once")
}

func (b *BaseBarFetcher) CurrentDateTime() *time.Time {
	return nil
}

func (b *BaseBarFetcher) ErrorC() <-chan error {
	return b.errorC
}

func (b *BaseBarFetcher) PendingBarsC() <-chan common.Bars {
	return b.pendingBars
}

func (b *BaseBarFetcher) IsRunning() bool {
	return !b.stopped
}
