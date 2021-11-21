package barfeed

import (
	"goalgotrade/common"
	"time"

	"github.com/go-gota/gota/series"
)

type MemBarFeed interface {
	common.BarFeed
}

type memBarFeed struct {
	baseBarFeed
}

func NewMemBarFeed(freqs []common.Frequency, stype series.Type, maxlen int) MemBarFeed {
	barfeed := NewBaseBarFeed(freqs, stype, maxlen)
	return &memBarFeed{
		baseBarFeed: *barfeed,
	}
}

func (m *memBarFeed) Reset() {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) GetCurrentDateTime() *time.Time {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) Start() error {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) Stop() error {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) Join() error {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) AddBarsFromSequence() error {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) Eof() bool {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) PeekDateTime() *time.Time {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) GetNextBars() common.Bars {
	// TODO: implement me
	panic("implement me")
}

func (m *memBarFeed) LoadAll() error {
	err := m.Start()
	if err != nil {
		return err
	}
	for {
		if m.Eof() {
			break
		}
		_, _, _, err := m.GetNextValuesAndUpdateDS()
		if err != nil {
			m.Stop()
			m.Join()
			return err
		}
	}
	return nil
}
