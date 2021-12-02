package broker

import (
	"fmt"
	orderconsts "goalgotrade/nugen/consts/order"
	"time"
)

// Order ...
type Order interface {
	Id() uint64
	IsActive() bool
	IsFilled() bool
	ExecutionInfo() OrderExecutionInfo
	AddExecutionInfo(info OrderExecutionInfo) error
	Remaining() int
	SwitchState(newState orderconsts.OrderState) error
}

// ValidTransitions ...
var ValidTransitions = map[orderconsts.OrderState][]orderconsts.OrderState{
	orderconsts.OrderStateINITIAL:          {orderconsts.OrderStateSUBMITTED, orderconsts.OrderStateCANCELED},
	orderconsts.OrderStateSUBMITTED:        {orderconsts.OrderStateACCEPTED, orderconsts.OrderStateCANCELED},
	orderconsts.OrderStateACCEPTED:         {orderconsts.OrderStatePARTIALLY_FILLED, orderconsts.OrderStateFILLED, orderconsts.OrderStateCANCELED},
	orderconsts.OrderStatePARTIALLY_FILLED: {orderconsts.OrderStatePARTIALLY_FILLED, orderconsts.OrderStateFILLED, orderconsts.OrderStateCANCELED},
}

// IsValidTransitions ...
func IsValidTransitions(from, to orderconsts.OrderState) bool {
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
	state         orderconsts.OrderState
	executionInfo OrderExecutionInfo

	quantity int
	filled   int
}

// NewOrder ...
func NewOrder() Order {
	return &order{}
}

// Id ...
func (o *order) Id() uint64 {
	return o.id
}

// IsActive ...
func (o *order) IsActive() bool {
	return o.state == orderconsts.OrderStateCANCELED || o.state == orderconsts.OrderStateFILLED
}

// IsFilled ...
func (o *order) IsFilled() bool {
	return o.state == orderconsts.OrderStateFILLED
}

// ExecutionInfo ...
func (o *order) ExecutionInfo() OrderExecutionInfo {
	return o.executionInfo
}

// Remaining ...
func (o *order) Remaining() int {
	return o.quantity - o.filled
}

// AddExecutionInfo ...
func (o *order) AddExecutionInfo(info OrderExecutionInfo) error {
	if info.Quantity > o.Remaining() {
		return fmt.Errorf("invalid fill size")
	}
	// TODO: implement me
	return nil
}

// SwitchState ...
func (o *order) SwitchState(newState orderconsts.OrderState) error {
	if IsValidTransitions(o.state, newState) {
		o.state = newState
		return nil
	}
	return fmt.Errorf("invalid state transition")
}

// NewOrderEvent ...
func NewOrderEvent(order Order, eventType orderconsts.OrderEventType, eventInfo OrderExecutionInfo) *OrderEvent {
	return &OrderEvent{
		Order:     order,
		EventType: eventType,
		EventInfo: eventInfo,
	}
}

// OrderEvent ...
type OrderEvent struct {
	Order     Order
	EventType orderconsts.OrderEventType
	EventInfo OrderExecutionInfo
}

// OrderExecutionInfo ...
type OrderExecutionInfo struct {
	Price      float64
	Quantity   int
	Commission float64
	Datetime   time.Time
	Info       string
}
