package feed

import (
	"fmt"
	"goalgotrade/nugen/bar"
	"goalgotrade/nugen/consts/frequency"
	"sort"
	"time"

	"github.com/go-gota/gota/series"
)

// MemBarFeed ...
type MemBarFeed interface {
	BaseBarFeed
	AddBarListFromSequence(instrument string, barList []bar.Bar) error
	LoadAll() error
}

type memBarFeed struct {
	baseBarFeed
	started         bool
	barData         map[string][]bar.Bar
	nextIdx         map[string]int
	currentDateTime *time.Time
}

// NewMemBarFeed ...
func NewMemBarFeed(freqList []frequency.Frequency, sType series.Type, maxLen int) MemBarFeed {
	return newMemBarFeed(freqList, sType, maxLen)
}

func newMemBarFeed(freqList []frequency.Frequency, sType series.Type, maxLen int) *memBarFeed {
	res := &memBarFeed{
		baseBarFeed: *newBaseBarFeed(freqList, sType, maxLen),
		barData:     map[string][]bar.Bar{},
		nextIdx:     map[string]int{},
	}
	return res
}

// Reset ...
func (m *memBarFeed) Reset(f BaseFeed) error {
	m.nextIdx = map[string]int{}
	for k := range m.barData {
		m.nextIdx[k] = 0
	}
	m.currentDateTime = nil
	return m.baseBarFeed.Reset(f)
}

// CurrentTime ...
func (m *memBarFeed) CurrentTime() *time.Time {
	return m.currentDateTime
}

// Start ...
func (m *memBarFeed) Start() error {
	if err := m.baseBarFeed.Start(); err != nil {
		return err
	}
	m.started = true
	return nil
}

// Stop ...
func (m *memBarFeed) Stop() error {
	// do nothing
	return nil
}

// Join ...
func (m *memBarFeed) Join() error {
	// do nothing
	return nil
}

// AddBarListFromSequence ...
func (m *memBarFeed) AddBarListFromSequence(instrument string, barList []bar.Bar) error {
	if m.started {
		return fmt.Errorf("can't add more bars once you started consuming bars")
	}

	if _, ok := m.nextIdx[instrument]; !ok {
		m.nextIdx[instrument] = 0
	}

	if _, ok := m.barData[instrument]; !ok {
		m.barData[instrument] = []bar.Bar{}
	}

	m.barData[instrument] = append(m.barData[instrument], barList...)
	if len(m.barData[instrument]) > 1 {
		sort.SliceStable(m.barData[instrument], func(i, j int) bool {
			return m.barData[instrument][i].Time().Before(*m.barData[instrument][j].Time())
		})
	}

	for _, bar := range barList {
		if err := m.RegisterInstrument(m, instrument, []frequency.Frequency{bar.Frequency()}); err != nil {
			return err
		}
	}
	return nil
}

// Eof ...
func (m *memBarFeed) Eof() bool {
	ret := true
	for _, instrument := range m.RegisteredInstruments() {
		barList := m.barData[instrument]
		if m.nextIdx[instrument] < len(barList) {
			ret = false
			break
		}
	}
	return ret
}

// PeekDateTime ...
func (m *memBarFeed) PeekDateTime() *time.Time {
	var resultDateTime *time.Time

	for _, instrument := range m.RegisteredInstruments() {
		nextIdx := m.nextIdx[instrument]
		barList := m.barData[instrument]
		if nextIdx < len(barList) {
			dateTime := barList[nextIdx].Time()
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

// NextBars ...
func (m *memBarFeed) NextBars() (bar.Bars, error) {
	smallestDateTime := m.PeekDateTime()
	if smallestDateTime == nil {
		return nil, fmt.Errorf("invalid datetime")
	}

	ret := bar.NewBars()
	for _, instrument := range m.RegisteredInstruments() {
		barList := m.barData[instrument]
		nextIdx := m.nextIdx[instrument]
		if nextIdx < len(barList) && barList[nextIdx].Time().Equal(*smallestDateTime) {
			if err := ret.AddBarList(instrument, barList); err != nil {
				return nil, err
			}
			m.nextIdx[instrument]++
		}
	}

	if m.currentDateTime.Equal(*smallestDateTime) {
		return nil, fmt.Errorf("duplicate bars found for %v on %s", ret.Instruments(), smallestDateTime)
	}

	m.currentDateTime = smallestDateTime
	return ret, nil
}

// LoadAll ...
func (m *memBarFeed) LoadAll() error {
	err := m.Start()
	if err != nil {
		return err
	}
	for {
		if m.Eof() {
			break
		}
		_, _, _, err := m.GetNextValuesAndUpdateDS(m)
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
