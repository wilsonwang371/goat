package broker

import (
	"fmt"
	orderconsts "goalgotrade/nugen/consts/order"
	"goalgotrade/nugen/core"
	"time"
)

// Broker ...
type Broker interface {
	core.Subject
	BrokerEssentials
}

// BrokerEssentials ...
type BrokerEssentials interface {
	NotifyOrderEvent(orderEvent *OrderEvent)
	OrderUpdatedChannel() core.Channel
}

type baseBroker struct {
	orderChannel core.Channel
	activeOrders map[uint64]Order
	feed         interface{}
}

// NewBrokerEssentials ...
func NewBrokerEssentials(feed core.Subject) BrokerEssentials {
	res := &baseBroker{
		orderChannel: core.NewChannel(),
		feed:         feed,
	}
	return res
}

// OrderUpdatedChannel ...
func (b *baseBroker) OrderUpdatedChannel() core.Channel {
	return b.orderChannel
}

// NotifyOrderEvent ...
func (b *baseBroker) NotifyOrderEvent(orderEvent *OrderEvent) {
	b.orderChannel.Emit(core.NewBasicEvent("order-event", map[string]interface{}{
		"event": orderEvent,
	}))
}

// CancelOrder ...
func (b *baseBroker) CancelOrder(order Order) error {
	if activeOrder, ok := b.activeOrders[order.Id()]; ok {
		if activeOrder.IsFilled() {
			return fmt.Errorf("can't cancel order that has already been filled")
		}
		b.unregisterOrder(activeOrder)
		activeOrder.SwitchState(orderconsts.OrderStateCANCELED)
		b.NotifyOrderEvent(NewOrderEvent(activeOrder, orderconsts.OrderEventCANCELED, OrderExecutionInfo{Info: "user requested cancellation"}))
	}
	return fmt.Errorf("the order is not active anymore")
}

func (b *baseBroker) unregisterOrder(order Order) error {
	if _, ok := b.activeOrders[order.Id()]; !ok {
		return fmt.Errorf("order not found")
	}
	delete(b.activeOrders, order.Id())
	return nil
}

// PeekDateTime ...
func (b *baseBroker) PeekDateTime() *time.Time {
	return nil
}

// Eof ...
func (b *baseBroker) Eof() bool {
	s := b.feed.(core.Subject)
	return s.Eof()
}
