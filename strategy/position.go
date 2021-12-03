package strategy

import (
	"fmt"
	"goalgotrade/broker"
	lg "goalgotrade/logger"
	"time"

	"go.uber.org/zap"
)

type enterExitOrder interface {
	EntryOrder() broker.Order
	ExitOrder() broker.Order
	EntryActive() bool
	EntryFilled() bool
	ExitActive() bool
	ExitFilled() bool
	SetEntryDateTime(dateTime time.Time)
	SetExitDateTime(dateTime time.Time)
}

// Position ...
type Position interface {
	enterExitOrder
	OnOrderEvent(orderEvent *broker.OrderEvent) error
	IsOpen(pos Position) bool
	BuildExitOrder(stopPrice, limitPrice float64) broker.Order
	Shares() int
	Strategy() Strategy
	SwitchState(pos Position, newState PositionState)
	GetActiveOrders() []broker.Order
}

// PositionState ...
type PositionState interface {
	CanSubmitOrder(position Position, order broker.Order) bool
	OnOrderEvent(position Position, orderEvent *broker.OrderEvent) error
	OnEnter(position Position) error
	IsOpen(position Position) bool
	Exit(position Position, stopPrice, limitPrice float64, goodTillCanceled bool) error
}

// PositionStateType ...
type PositionStateType int

// PositionStateWaitingEntryState ...
const (
	PositionStateWaitingEntryState PositionStateType = iota
	PositionStateOpenState
	PositionStateClosedState
)

type basePosition struct {
	state PositionState

	entryOrder    broker.Order
	entryDateTime *time.Time
	exitOrder     broker.Order
	exitDateTime  *time.Time
	allOrNone     bool
	activeOrders  map[int]broker.Order

	strategy Strategy
	shares   int
}

// NewBasePosition ...
func NewBasePosition(strategy Strategy, entryOrder broker.Order, goodTillCanceled, allOrNone bool) Position {
	return newBasePosition(strategy, entryOrder, goodTillCanceled, allOrNone)
}

func newBasePosition(strategy Strategy, entryOrder broker.Order, goodTillCanceled, allOrNone bool) *basePosition {
	res := &basePosition{
		strategy:     strategy,
		entryOrder:   entryOrder,
		allOrNone:    allOrNone,
		activeOrders: map[int]broker.Order{},
	}
	// TODO: implement me
	return res
}

// OnOrderEvent ...
func (p *basePosition) OnOrderEvent(orderEvent *broker.OrderEvent) error {
	return nil
}

// IsOpen ...
func (p *basePosition) IsOpen(pos Position) bool {
	return p.state.IsOpen(pos)
}

// EntryActive ...
func (p *basePosition) EntryActive() bool {
	return p.entryOrder != nil && p.entryOrder.IsActive()
}

// ExitActive ...
func (p *basePosition) ExitActive() bool {
	return p.exitOrder != nil && p.exitOrder.IsActive()
}

// EntryFilled ...
func (p *basePosition) EntryFilled() bool {
	return p.exitOrder != nil && p.entryOrder.IsFilled()
}

// ExitFilled ...
func (p *basePosition) ExitFilled() bool {
	return p.exitOrder != nil && p.exitOrder.IsFilled()
}

// EntryOrder ...
func (p *basePosition) EntryOrder() broker.Order {
	return p.entryOrder
}

// ExitOrder ...
func (p *basePosition) ExitOrder() broker.Order {
	return p.exitOrder
}

// SetEntryDateTime ...
func (p *basePosition) SetEntryDateTime(dateTime time.Time) {
	tmpTime := dateTime
	p.entryDateTime = &tmpTime
}

// SetExitDateTime ...
func (p *basePosition) SetExitDateTime(dateTime time.Time) {
	tmpTime := dateTime
	p.exitDateTime = &tmpTime
}

// SwitchState ...
func (p *basePosition) SwitchState(pos Position, newState PositionState) {
	p.state = newState
	if err := p.state.OnEnter(pos); err != nil {
		lg.Logger.Warn("switch state failed", zap.Error(err))
	}
}

// Strategy ...
func (p *basePosition) Strategy() Strategy {
	return p.strategy
}

// Shares ...
func (p *basePosition) Shares() int {
	return p.shares
}

func (p *basePosition) submitExitOrder(stopPrice, limitPrice float64, goodTillCanceled bool) error {
	// TODO: implement me
	return nil
}

// BuildExitOrder ...
func (p *basePosition) BuildExitOrder(stopPrice, limitPrice float64) broker.Order {
	panic("not implemented")
	return nil
}

// GetAge ...
func (p *basePosition) GetAge() *time.Duration {
	if p.entryDateTime != nil {
		var res time.Duration
		if p.exitDateTime != nil {
			res = p.exitDateTime.Sub(*p.exitDateTime)
		} else {
			tmp := p.strategy.CurrentTime()
			if tmp == nil {
				lg.Logger.Warn("empty strategy GetCurrentDateTime")
				return nil
			}
			res = tmp.Sub(*p.exitDateTime)
		}
		return &res
	}
	lg.Logger.Warn("empty entry time")
	return nil
}

// GetActiveOrders ...
func (p *basePosition) GetActiveOrders() []broker.Order {
	var res []broker.Order
	for _, order := range p.activeOrders {
		res = append(res, order)
	}
	return res
}

// NewPositionState ...
func NewPositionState(stateType PositionStateType) PositionState {
	switch stateType {
	case PositionStateWaitingEntryState:
		return &WaitingEntryState{}
	case PositionStateOpenState:
		return &OpenState{}
	case PositionStateClosedState:
		return &ClosedState{}
	}
	return nil
}

// WaitingEntryState ...
type WaitingEntryState struct{}

// CanSubmitOrder ...
func (w *WaitingEntryState) CanSubmitOrder(position Position, order broker.Order) bool {
	if position.EntryActive() {
		return false
	}
	return true
}

// OnOrderEvent ...
func (w *WaitingEntryState) OnOrderEvent(position Position, orderEvent *broker.OrderEvent) error {
	// TODO: implement me
	return nil
}

// OnEnter ...
func (w *WaitingEntryState) OnEnter(position Position) error {
	return nil
}

// IsOpen ...
func (w *WaitingEntryState) IsOpen(position Position) bool {
	return true
}

// Exit ...
func (w *WaitingEntryState) Exit(position Position, stopPrice, limitPrice float64, goodTillCanceled bool) error {
	if position.Shares() == 0 {
		return fmt.Errorf("no shares")
	}
	if !position.EntryOrder().IsActive() {
		return fmt.Errorf("entry order is not active")
	}
	err := position.Strategy().Broker().CancelOrder(position.EntryOrder())
	if err != nil {
		return err
	}
	return nil
}

// OpenState ...
type OpenState struct{}

// CanSubmitOrder ...
func (o *OpenState) CanSubmitOrder(position Position, order broker.Order) bool {
	return true
}

// OnOrderEvent ...
func (o *OpenState) OnOrderEvent(position Position, orderEvent *broker.OrderEvent) error {
	// TODO: Implement me
	return nil
}

// OnEnter ...
func (o *OpenState) OnEnter(position Position) error {
	entryDateTime := position.EntryOrder().ExecutionInfo().Datetime
	position.SetEntryDateTime(entryDateTime)
	return nil
}

// IsOpen ...
func (o *OpenState) IsOpen(position Position) bool {
	return true
}

// Exit ...
func (o *OpenState) Exit(pos Position, stopPrice, limitPrice float64, goodTillCanceled bool) error {
	if pos.Shares() == 0 {
		return fmt.Errorf("no shares")
	}
	if pos.ExitActive() {
		return fmt.Errorf("exit oder is active and it should be cancelled first")
	}
	if pos.EntryActive() {
		err := pos.Strategy().Broker().CancelOrder(pos.EntryOrder())
		if err != nil {
			return err
		}
	}
	if pos2, ok := pos.(*basePosition); ok {
		if err := pos2.submitExitOrder(stopPrice, limitPrice, goodTillCanceled); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("failed to submit exit order")
	}

	return nil
}

// ClosedState ...
type ClosedState struct{}

// CanSubmitOrder ...
func (c *ClosedState) CanSubmitOrder(position Position, order broker.Order) bool {
	return false
}

// OnOrderEvent ...
func (c *ClosedState) OnOrderEvent(position Position, orderEvent *broker.OrderEvent) error {
	return nil
}

// OnEnter ...
func (c *ClosedState) OnEnter(position Position) error {
	if position.ExitFilled() {
		exitDateTime := position.ExitOrder().ExecutionInfo().Datetime
		position.SetExitDateTime(exitDateTime)
	}
	if position.Shares() == 0 {
		return fmt.Errorf("no shares")
	}
	if err := position.Strategy().UnregisterPosition(position); err != nil {
		return err
	}
	return nil
}

// IsOpen ...
func (c *ClosedState) IsOpen(position Position) bool {
	return false
}

// Exit ...
func (c *ClosedState) Exit(position Position, stopPrice, limitPrice float64, goodTillCanceled bool) error {
	return nil
}

// LongPosition ...
type LongPosition struct {
	basePosition
}

// NewLongPosition ...
func NewLongPosition(stopPrice, limitPrice float64) Position {
	// TODO: implement me
	return nil
}

// ShortPosition ...
type ShortPosition struct {
	basePosition
}

// NewShortPosition ...
func NewShortPosition(stopPrice, limitPrice float64) Position {
	// TODO: implement me
	return nil
}
