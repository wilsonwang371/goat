package broker

import "goalgotrade/common"

func NewOrderEvent(order common.Order, eventType common.OrderEventType, eventInfo common.OrderExecutionInfo) *common.OrderEvent {
	return &common.OrderEvent{
		Order:     order,
		EventType: eventType,
		EventInfo: eventInfo,
	}
}
