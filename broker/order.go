package broker

import (
	"fmt"
	"goalgotrade/common"
)

var ValidTransitions = map[common.OrderState][]common.OrderState{
	common.OrderState_INITIAL:          {common.OrderState_SUBMITTED, common.OrderState_CANCELED},
	common.OrderState_SUBMITTED:        {common.OrderState_ACCEPTED, common.OrderState_CANCELED},
	common.OrderState_ACCEPTED:         {common.OrderState_PARTIALLY_FILLED, common.OrderState_FILLED, common.OrderState_CANCELED},
	common.OrderState_PARTIALLY_FILLED: {common.OrderState_PARTIALLY_FILLED, common.OrderState_FILLED, common.OrderState_CANCELED},
}

func IsValidTransitions(from, to common.OrderState) bool {
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
	state         common.OrderState
	executionInfo common.OrderExecutionInfo

	quantity int
	filled   int
}

func NewOrder() common.Order {
	return &order{}
}

func (o *order) GetId() uint64 {
	return o.id
}

func (o *order) IsActive() bool {
	return o.state == common.OrderState_CANCELED || o.state == common.OrderState_FILLED
}

func (o *order) IsFilled() bool {
	return o.state == common.OrderState_FILLED
}

func (o *order) GetExecutionInfo() common.OrderExecutionInfo {
	return o.executionInfo
}

func (o *order) GetRemaining() int {
	return o.quantity - o.filled
}

func (o *order) AddExecutionInfo(info common.OrderExecutionInfo) error {
	if info.Quantity > o.GetRemaining() {
		return fmt.Errorf("invalid fill size")
	}
	// TODO: implement me
	return nil
}

func (o *order) SwitchState(newState common.OrderState) error {
	if IsValidTransitions(o.state, newState) {
		o.state = newState
		return nil
	}
	return fmt.Errorf("invalid state transition")
}
