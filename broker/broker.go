package broker

import (
	"fmt"
	"goalgotrade/common"
	"goalgotrade/core"
	"time"
)

type broker struct {
	*core.DefaultSubject
	orderEvent   common.Event
	activeOrders map[uint64]common.Order
	barfeed      common.BarFeed
}

func NewBroker(barfeed common.BarFeed) common.Broker {
	return &broker{
		DefaultSubject: core.NewDefaultSubject(),
		orderEvent:     core.NewEvent(),
		barfeed:        barfeed,
	}
}

func (b *broker) GetOrderUpdatedEvent() common.Event {
	return b.orderEvent
}

func (b *broker) NotifyOrderEvent(orderEvent *common.OrderEvent) {
	b.orderEvent.Emit(orderEvent)
}

func (b *broker) CancelOrder(order common.Order) error {
	if activeOrder, ok := b.activeOrders[order.GetId()]; ok {
		if activeOrder.IsFilled() {
			return fmt.Errorf("can't cancel order that has already been filled")
		}
		b.unregisterOrder(activeOrder)
		activeOrder.SwitchState(common.OrderState_CANCELED)
		b.NotifyOrderEvent(NewOrderEvent(activeOrder, common.OrderEventType_CANCELED, common.OrderExecutionInfo{Info: "user requested cancellation"}))
	}
	return fmt.Errorf("the order is not active anymore")
}

func (b *broker) unregisterOrder(order common.Order) error {
	if _, ok := b.activeOrders[order.GetId()]; !ok {
		return fmt.Errorf("order not found")
	}
	delete(b.activeOrders, order.GetId())
	return nil
}

func (b *broker) PeekDateTime() *time.Time {
	return nil
}

func (b *broker) Eof() bool {
	return b.barfeed.Eof()
}
