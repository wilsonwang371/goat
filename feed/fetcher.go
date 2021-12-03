package feed

import (
	"fmt"
	"goalgotrade/bar"
	"goalgotrade/consts/frequency"
	lg "goalgotrade/logger"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/go-gota/gota/series"
)

// DefaultDataPullInterval ...
const DefaultDataPullInterval = 10 * time.Second

type baseBarFetcher struct {
	mu           sync.Mutex
	instrument   string
	freqList     []frequency.Frequency
	pendingBars  chan bar.Bars
	stopped      bool
	stopC        chan struct{}
	doneC        chan struct{}
	errorC       chan error
	pullInterval time.Duration
	provider     BarFetcherProvider
}

type BarFetcherProvider interface {
	init(instrument string, freqList []frequency.Frequency) error
	connect() error
	nextBars() (bar.Bars, error)
	reset() error
	stop() error
	datatype() series.Type
}

// NewBaseBarFetcher ...
func NewBaseBarFetcher(provider BarFetcherProvider, pullInterval time.Duration) BarFetcher {
	return newBaseBarFetcher(provider, pullInterval)
}

func newBaseBarFetcher(provider BarFetcherProvider, pullInterval time.Duration) *baseBarFetcher {
	b := &baseBarFetcher{
		instrument:   "",
		pendingBars:  make(chan bar.Bars, 1024),
		stopped:      true,
		stopC:        make(chan struct{}, 1),
		doneC:        make(chan struct{}, 1),
		errorC:       make(chan error, 8),
		pullInterval: DefaultDataPullInterval,
		provider:     provider,
	}
	if pullInterval >= 0 {
		b.pullInterval = pullInterval
	}
	return b
}

// Start ...
func (b *baseBarFetcher) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.stopped {
		return fmt.Errorf("already running")
	}
	if b.instrument == "" || len(b.freqList) == 0 {
		return fmt.Errorf("no instrument registered")
	}
	b.stopped = false
	if err := b.provider.init(b.instrument, b.freqList); err != nil {
		b.stopped = true
		return err
	}
	go b.run()
	return nil
}

func (b *baseBarFetcher) appendError(err error) error {
	select {
	case b.errorC <- err:
	default:
		return fmt.Errorf("failed to append error")
	}
	return nil
}

func (b *baseBarFetcher) run() {
	defer func() {
		close(b.doneC)
		b.stopped = true
	}()
	t := time.NewTimer(b.pullInterval)
	if err := b.provider.connect(); err != nil {
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

		if bars, err := b.provider.nextBars(); err != nil {
			lg.Logger.Error("nextBars failed", zap.Error(err))
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

// Stop ...
func (b *baseBarFetcher) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.stopped {
		return fmt.Errorf("already stopped")
	}
	if err := b.provider.stop(); err != nil {
		return err
	}
	close(b.stopC)
	<-b.doneC
	return nil
}

// RegisterInstrument ...
func (b *baseBarFetcher) RegisterInstrument(instrument string, freqList []frequency.Frequency) error {
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

// CurrentDateTime ...
func (b *baseBarFetcher) CurrentDateTime() *time.Time {
	return nil
}

// ErrorC ...
func (b *baseBarFetcher) ErrorC() <-chan error {
	return b.errorC
}

// PendingBarsC ...
func (b *baseBarFetcher) PendingBarsC() <-chan bar.Bars {
	return b.pendingBars
}

// IsRunning ...
func (b *baseBarFetcher) IsRunning() bool {
	return !b.stopped
}

// GetInstrument ...
func (b *baseBarFetcher) GetInstrument() string {
	return b.instrument
}

// GetFrequencies ...
func (b *baseBarFetcher) GetFrequencies() []frequency.Frequency {
	return b.freqList
}

// GetDSType ...
func (b *baseBarFetcher) GetDSType() series.Type {
	return b.provider.datatype()
}
