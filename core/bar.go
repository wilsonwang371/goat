package core

import (
	"fmt"
	"time"

	"goalgotrade/common"
)

type basicBar struct {
}

func NewBasicBar() common.Bar {
	return &basicBar{}
}

func (b *basicBar) SetUseAdjustedValue(useAdjusted bool) error {
	return fmt.Errorf("implement me")
}

func (b *basicBar) GetUseAdjValue() bool {
	// TODO: implement me
	return false
}

func (b *basicBar) GetDateTime() time.Time {
	// TODO: implement me
	return time.Now()
}

func (b *basicBar) Open(adjusted bool) float64 {
	// TODO: implement me
	return .0
}

func (b *basicBar) High(adjusted bool) float64 {
	// TODO: implement me
	return .0
}

func (b *basicBar) Low(adjusted bool) float64 {
	// TODO: implement me
	return .0
}

func (b *basicBar) Close(adjusted bool) float64 {
	// TODO: implement me
	return .0
}

func (b *basicBar) Volume() int {
	// TODO: implement me
	return 0
}

func (b *basicBar) AdjClose() float64 {
	// TODO: implement me
	return .0
}

func (b *basicBar) Frequency() float64 {
	// TODO: implement me
	return .0
}

func (b *basicBar) Price() float64 {
	// TODO: implement me
	return .0
}
