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
	SetMeta(key int, value interface{})
	GetMeta(key int) interface{}
}

type BasicBarData struct {
	OpenV        float64             `json:"open"`
	HighV        float64             `json:"high"`
	LowV         float64             `json:"low"`
	CloseV       float64             `json:"close"`
	VolumeV      int64               `json:"volume"`
	AdjCloseV    float64             `json:"adjClose"`
	FrequencyV   Frequency           `json:"frequency"`
	DateTimeV    time.Time           `json:"dateTime"`
	UseAdjustedV bool                `json:"useAdjusted"`
	Meta         map[int]interface{} `json:"meta"`
}

const (
	BarMetaIsRecovery = iota + 1
	BarMetaEnd
)

// String implements Bar
func (b *BasicBarData) String() string {
	return fmt.Sprintf("%s %f %f %f %f %d", b.DateTimeV.Format("2006-01-02 15:04:05"),
		b.OpenV, b.HighV, b.LowV, b.CloseV, b.VolumeV)
}

// AdjClose implements Bar
func (b *BasicBarData) AdjClose() float64 {
	return b.AdjCloseV
}

// Close implements Bar
func (b *BasicBarData) Close() float64 {
	return b.CloseV
}

// DateTime implements Bar
func (b *BasicBarData) DateTime() time.Time {
	return b.DateTimeV
}

// Frequency implements Bar
func (b *BasicBarData) Frequency() Frequency {
	return b.FrequencyV
}

// High implements Bar
func (b *BasicBarData) High() float64 {
	return b.HighV
}

// Low implements Bar
func (b *BasicBarData) Low() float64 {
	return b.LowV
}

// Open implements Bar
func (b *BasicBarData) Open() float64 {
	return b.OpenV
}

// SetUseAdjustedValue implements Bar
func (b *BasicBarData) SetUseAdjustedValue(v bool) {
	b.UseAdjustedV = v
}

// Volume implements Bar
func (b *BasicBarData) Volume() int64 {
	return b.VolumeV
}

func (b *BasicBarData) SetMeta(key int, value interface{}) {
	if key <= 0 || key >= BarMetaEnd {
		panic("invalid key")
	}
	if b.Meta == nil {
		b.Meta = make(map[int]interface{})
	}
	b.Meta[key] = value
}

func (b *BasicBarData) GetMeta(key int) interface{} {
	if key <= 0 || key >= BarMetaEnd {
		panic("invalid key")
	}
	if b.Meta == nil {
		return nil
	}
	if v, ok := b.Meta[key]; ok {
		return v
	}
	return nil
}

func NewBasicBar(t time.Time, open, high, low, close, adj_close float64, volume int64, frequency Frequency) Bar {
	return &BasicBarData{
		OpenV:      open,
		HighV:      high,
		LowV:       low,
		CloseV:     close,
		VolumeV:    volume,
		FrequencyV: frequency,
		DateTimeV:  t,
		Meta:       make(map[int]interface{}),
	}
}
