package broker

import (
	"fmt"
)

type Order interface {
	GetId() uint64
	IsActive() bool
	IsFilled() bool
	GetExecutionInfo() OrderExecutionInfo
	AddExecutionInfo(info OrderExecutionInfo) error

	GetRemaining() int

	SwitchState(newState OrderState) error
}

type OrderState int

const (
	OS_UNKNOWN          OrderState = iota
	OS_INITIAL                     // Initial state.
	OS_SUBMITTED                   // Order has been submitted.
	OS_ACCEPTED                    // Order has been acknowledged by the broker.
	OS_CANCELED                    // Order has been canceled.
	OS_PARTIALLY_FILLED            // Order has been partially filled.
	OS_FILLED                      // Order has been completely filled.
)

var ValidTransitions = map[OrderState][]OrderState{
	OS_INITIAL:          {OS_SUBMITTED, OS_CANCELED},
	OS_SUBMITTED:        {OS_ACCEPTED, OS_CANCELED},
	OS_ACCEPTED:         {OS_PARTIALLY_FILLED, OS_FILLED, OS_CANCELED},
	OS_PARTIALLY_FILLED: {OS_PARTIALLY_FILLED, OS_FILLED, OS_CANCELED},
}

func IsValidTransitions(from, to OrderState) bool {
	if l, ok := ValidTransitions[from]; ok {
		for _, v := range l {
			if v == to {
				return true
			}
		}
	}
	return false
}

type order struct {
	id            uint64
	state         OrderState
	executionInfo OrderExecutionInfo

	quantity int
	filled   int
}

func NewOrder() Order {
	return &order{}
}

func (o *order) GetId() uint64 {
	return o.id
}

func (o *order) IsActive() bool {
	return o.state == OS_CANCELED || o.state == OS_FILLED
}

func (o *order) IsFilled() bool {
	return o.state == OS_FILLED
}

func (o *order) GetExecutionInfo() OrderExecutionInfo {
	return o.executionInfo
}

func (o *order) GetRemaining() int {
	return o.quantity - o.filled
}

func (o *order) AddExecutionInfo(info OrderExecutionInfo) error {
	if info.Quantity > o.GetRemaining() {
		return fmt.Errorf("invalid fill size")
	}
	// TODO: implement me
	return nil
}

func (o *order) SwitchState(newState OrderState) error {
	if IsValidTransitions(o.state, newState) {
		o.state = newState
		return nil
	}
	return fmt.Errorf("invalid state transition")
}
