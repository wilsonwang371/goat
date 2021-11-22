package barfeed

import (
	"fmt"
	"goalgotrade/common"
	"goalgotrade/core"
	"sort"
	"time"

	"github.com/go-gota/gota/series"
)

type MemBarFeed interface {
	common.BarFeed
	AddBarListFromSequence(instrument string, barList []common.Bar) error
	LoadAll() error
}

type memBarFeed struct {
	baseBarFeed
	started         bool
	bars            common.Bars
	nextIdx         map[string]int
	currentDateTime *time.Time
}

func NewMemBarFeed(freqList []common.Frequency, sType series.Type, maxLen int) *memBarFeed {
	barFeed := NewBaseBarFeed(freqList, sType, maxLen)
	return &memBarFeed{
		baseBarFeed: *barFeed,
		bars:        core.NewBars(),
		nextIdx:     map[string]int{},
	}
}

func (m *memBarFeed) Reset() {
	m.nextIdx = map[string]int{}
	for _, instrument := range m.bars.GetInstruments() {
		m.nextIdx[instrument] = 0
	}
	m.currentDateTime = nil
	m.baseBarFeed.Reset()
}

func (m *memBarFeed) GetCurrentDateTime() *time.Time {
	return m.currentDateTime
}

func (m *memBarFeed) Start() error {
	if err := m.baseBarFeed.Start(); err != nil {
		return err
	}
	m.started = true
	return nil
}

func (m *memBarFeed) Stop() error {
	// do nothing
	return nil
}

func (m *memBarFeed) Join() error {
	// do nothing
	return nil
}

func (m *memBarFeed) AddBarListFromSequence(instrument string, barList []common.Bar) error {
	if m.started {
		return fmt.Errorf("can't add more bars once you started consuming bars")
	}

	if _, ok := m.nextIdx[instrument]; !ok {
		m.nextIdx[instrument] = 0
	}

	if err := m.bars.AddBarList(instrument, barList); err != nil {
		return err
	}
	newBarList := m.bars.GetBarList(instrument)
	if len(newBarList) > 1 {
		sort.SliceStable(newBarList, func(i, j int) bool {
			return newBarList[i].GetDateTime().Before(*newBarList[j].GetDateTime())
		})
	}
	allFreq := map[common.Frequency]bool{}
	for _, v := range barList {
		if _, ok := allFreq[v.Frequency()]; !ok {
			allFreq[v.Frequency()] = true
		}
	}
	for freq := range allFreq {
		if err := m.RegisterInstrument(instrument, freq); err != nil {
			return err
		}
	}
	return nil
}

func (m *memBarFeed) Eof() bool {
	ret := true
	for _, instrument := range m.bars.GetInstruments() {
		barList := m.bars.GetBarList(instrument)
		if m.nextIdx[instrument] < len(barList) {
			ret = false
			break
		}
	}
	return ret
}

func (m *memBarFeed) PeekDateTime() *time.Time {
	var resultDateTime *time.Time

	for _, instrument := range m.bars.GetInstruments() {
		nextIdx := m.nextIdx[instrument]
		barList := m.bars.GetBarList(instrument)
		if nextIdx < len(barList) {
			dateTime := barList[nextIdx].GetDateTime()
			if resultDateTime == nil {
				if dateTime != nil {
					resultDateTime = dateTime
				}
			} else {
				if dateTime != nil && dateTime.Before(*resultDateTime) {
					resultDateTime = dateTime
				}
			}
		}
	}
	return resultDateTime
}

func (m *memBarFeed) GetNextBars() (common.Bars, error) {
	smallestDateTime := m.PeekDateTime()
	if smallestDateTime == nil {
		return nil, fmt.Errorf("invalid datetime")
	}

	ret := core.NewBars()
	for _, instrument := range m.bars.GetInstruments() {
		barList := m.bars.GetBarList(instrument)
		nextIdx := m.nextIdx[instrument]
		if nextIdx < len(barList) && barList[nextIdx].GetDateTime().Equal(*smallestDateTime) {
			if err := ret.AddBarList(instrument, barList); err != nil {
				return nil, err
			}
			m.nextIdx[instrument]++
		}
	}

	if m.currentDateTime.Equal(*smallestDateTime) {
		return nil, fmt.Errorf("duplicate bars found for %v on %s", ret.GetInstruments(), smallestDateTime)
	}

	m.currentDateTime = smallestDateTime
	return ret, nil
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
			if err := m.Stop(); err != nil {
				return err
			}
			if err := m.Join(); err != nil {
				return err
			}
			return err
		}
	}
	return nil
}
