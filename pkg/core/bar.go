package core

import (
	"fmt"
	"time"
)

type Bars map[string]Bar

type Bar interface {
	Open() float64
	High() float64
	Low() float64
	Close() float64
	Volume() int64
	AdjClose() float64
	Frequency() Frequency
	DateTime() time.Time
	SetUseAdjustedValue(bool)
	String() string
}

type bar struct {
	open        float64   `json:"open"`
	high        float64   `json:"high"`
	low         float64   `json:"low"`
	close       float64   `json:"close"`
	volume      int64     `json:"volume"`
	adjClose    float64   `json:"adjClose"`
	frequency   Frequency `json:"frequency"`
	dateTime    time.Time `json:"dateTime"`
	useAdjusted bool      `json:"useAdjusted"`
}

// String implements Bar
func (b *bar) String() string {
	return fmt.Sprintf("%s %f %f %f %f %d", b.dateTime.Format("2006-01-02 15:04:05"),
		b.open, b.high, b.low, b.close, b.volume)
}

// AdjClose implements Bar
func (b *bar) AdjClose() float64 {
	return b.adjClose
}

// Close implements Bar
func (b *bar) Close() float64 {
	return b.close
}

// DateTime implements Bar
func (b *bar) DateTime() time.Time {
	return b.dateTime
}

// Frequency implements Bar
func (b *bar) Frequency() Frequency {
	return b.frequency
}

// High implements Bar
func (b *bar) High() float64 {
	return b.high
}

// Low implements Bar
func (b *bar) Low() float64 {
	return b.low
}

// Open implements Bar
func (b *bar) Open() float64 {
	return b.open
}

// SetUseAdjustedValue implements Bar
func (b *bar) SetUseAdjustedValue(v bool) {
	b.useAdjusted = v
}

// Volume implements Bar
func (b *bar) Volume() int64 {
	return b.volume
}

func NewBasicBar(t time.Time, open, high, low, close, adj_close float64, volume int64, frequency Frequency) Bar {
	return &bar{
		open:      open,
		high:      high,
		low:       low,
		close:     close,
		volume:    volume,
		frequency: frequency,
		dateTime:  t,
	}
}
