package core

import "time"

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
}

type bar struct {
	open        float64
	high        float64
	low         float64
	close       float64
	volume      int64
	adjClose    float64
	frequency   Frequency
	dateTime    time.Time
	useAdjusted bool
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

func NewBasicBar(open, high, low, close float64, volume int64, frequency Frequency, t time.Time) Bar {
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
