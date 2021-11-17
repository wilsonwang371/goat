package common

import (
	"time"

	"github.com/go-gota/gota/series"
)

type OrderExecutionInfo struct {
	Price      float64
	Quantity   int
	Commission float64
	Datetime   time.Time
	Info       string
}

func NewOrderEvent(order Order, eventType OrderEventType, eventInfo OrderExecutionInfo) *OrderEvent {
	return &OrderEvent{
		Order:     order,
		EventType: eventType,
		EventInfo: eventInfo,
	}
}

type OrderEvent struct {
	Order     Order
	EventType OrderEventType
	EventInfo OrderExecutionInfo
}

type OrderEventType int

type OrderState int

type EventHandler func(args ...interface{}) error

type DataSeriesType series.Type
