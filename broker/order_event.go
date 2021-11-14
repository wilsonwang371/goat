package broker

import "time"

type OrderEvent struct {
	order     Order
	eventType OrderEventType
	eventInfo OrderExecutionInfo
}

type OrderEventType int

const (
	OET_SUBMITTED        OrderEventType = iota + 1 // Order has been submitted.
	OET_ACCEPTED                                   // Order has been acknowledged by the broker.
	OET_CANCELED                                   // Order has been canceled.
	OET_PARTIALLY_FILLED                           // Order has been partially filled.
	OET_FILLED                                     // Order has been completely filled.
)

func NewOrderEvent(order Order, eventType OrderEventType, eventInfo OrderExecutionInfo) *OrderEvent {
	return &OrderEvent{
		order:     order,
		eventType: eventType,
		eventInfo: eventInfo,
	}
}

func (e *OrderEvent) GetOrder() Order {
	return e.order
}

func (e *OrderEvent) GetEventType() OrderEventType {
	return e.eventType
}

func (e *OrderEvent) GetEventInfo() OrderExecutionInfo {
	return e.eventInfo
}

type OrderExecutionInfo struct {
	Price      float64
	Quantity   int
	Commission float64
	Datetime   time.Time
	Info       string
}
