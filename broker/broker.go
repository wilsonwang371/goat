package broker

import (
	"fmt"

	"goalgotrade/common"
	"goalgotrade/core"
)

type Broker interface {
	common.Subject
	GetOrderUpdatedEvent() common.Event
	NotifyOrderEvent(orderEvent *OrderEvent)
	CancelOrder(order Order) error
}

type broker struct {
	*core.DefaultSubject
	orderEvent   common.Event
	activeOrders map[uint64]Order
}

func NewBroker() Broker {
	return &broker{
		DefaultSubject: core.NewDefaultSubject(),
		orderEvent:     core.NewEvent(),
	}
}

func (b *broker) GetOrderUpdatedEvent() common.Event {
	return b.orderEvent
}

func (b *broker) NotifyOrderEvent(orderEvent *OrderEvent) {
	b.orderEvent.Emit(orderEvent)
}

func (b *broker) CancelOrder(order Order) error {
	if activeOrder, ok := b.activeOrders[order.GetId()]; ok {
		if activeOrder.IsFilled() {
			return fmt.Errorf("can't cancel order that has already been filled")
		}
		b.unregisterOrder(activeOrder)
		activeOrder.SwitchState(OS_CANCELED)
		b.NotifyOrderEvent(NewOrderEvent(activeOrder, OET_CANCELED, OrderExecutionInfo{Info: "user requested cancellation"}))
	}
	return fmt.Errorf("the order is not active anymore")
}

func (b *broker) unregisterOrder(order Order) error {
	if _, ok := b.activeOrders[order.GetId()]; !ok {
		return fmt.Errorf("order not found")
	}
	delete(b.activeOrders, order.GetId())
	return nil
}
