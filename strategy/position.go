package strategy

import (
	"fmt"
	"time"

	"goalgotrade/common"
)

type Position interface {
	OnOrderEvent(orderEvent *common.OrderEvent) error
	IsOpen() bool

	GetEntryOrder() common.Order
	EntryActive() bool
	EntryFilled() bool
	SetEntryDateTime(datetime time.Time)

	GetExitOrder() common.Order
	ExitActive() bool
	ExitFilled() bool
	SetExitDateTime(datetime time.Time)

	GetShares() int
	GetStrategy() Strategy
}

type PositionState interface {
	CanSubmitOrder(position Position, order common.Order) bool
	OnOrderEvent(position Position, orderEvent *common.OrderEvent) error
	OnEnter(position Position) error
	IsOpen(position Position) bool
	Exit(position Position, stopPrice, limitPrice float64, goodTillCanceled bool) error
}

type PositionStateType int

const (
	PS_WaitingEntryState PositionStateType = iota
	PS_OpenState
	PS_ClosedState
)

type position struct {
	state PositionState

	entryOrder common.Order
	exitOrder  common.Order

	entryDateTime *time.Time
	exitDateTime  *time.Time

	strategy Strategy
	shares   int
}

func NewPosition(strategy Strategy) Position {
	return &position{
		strategy: strategy,
	}
}

func (p *position) OnOrderEvent(orderEvent *common.OrderEvent) error {
	return nil
}

func (p *position) IsOpen() bool {
	return p.state.IsOpen(p)
}

func (p *position) EntryActive() bool {
	return p.entryOrder != nil && p.entryOrder.IsActive()
}

func (p *position) ExitActive() bool {
	return p.exitOrder != nil && p.exitOrder.IsActive()
}

func (p *position) EntryFilled() bool {
	return p.exitOrder != nil && p.entryOrder.IsFilled()
}

func (p *position) ExitFilled() bool {
	return p.exitOrder != nil && p.exitOrder.IsFilled()
}

func (p *position) GetEntryOrder() common.Order {
	return p.entryOrder
}

func (p *position) GetExitOrder() common.Order {
	return p.exitOrder
}

func (p *position) SetEntryDateTime(datetime time.Time) {
	tmptime := datetime
	p.entryDateTime = &tmptime
}

func (p *position) SetExitDateTime(datetime time.Time) {
	tmptime := datetime
	p.exitDateTime = &tmptime
}

func (p *position) GetStrategy() Strategy {
	return p.strategy
}

func (p *position) GetShares() int {
	return p.shares
}

func (p *position) submitExitOrder(stopPrice, limitPrice float64, goodTillCanceled bool) error {
	// TODO: implement me
	return nil
}

func NewPositionState(stateType PositionStateType) PositionState {
	switch stateType {
	case PS_WaitingEntryState:
		return &WaitingEntryState{}
	case PS_OpenState:
		return &OpenState{}
	case PS_ClosedState:
		return &ClosedState{}
	}
	return nil
}

type WaitingEntryState struct {
}

func (w *WaitingEntryState) CanSubmitOrder(position Position, order common.Order) bool {
	if position.EntryActive() {
		return false
	}
	return true
}

func (w *WaitingEntryState) OnOrderEvent(position Position, orderEvent *common.OrderEvent) error {
	// TODO: implement me
	return nil
}

func (w *WaitingEntryState) OnEnter(position Position) error {
	return nil
}

func (w *WaitingEntryState) IsOpen(position Position) bool {
	return true
}

func (w *WaitingEntryState) Exit(position Position, stopPrice, limitPrice float64, goodTillCanceled bool) error {
	if position.GetShares() == 0 {
		return fmt.Errorf("no shares")
	}
	if !position.GetEntryOrder().IsActive() {
		return fmt.Errorf("entry order is not active")
	}
	position.GetStrategy().GetBroker().CancelOrder(position.GetEntryOrder())
	return nil
}

type OpenState struct {
}

func (o *OpenState) CanSubmitOrder(position Position, order common.Order) bool {
	return true
}

func (o *OpenState) OnOrderEvent(position Position, orderEvent *common.OrderEvent) error {
	// TODO: Implement me
	return nil
}

func (o *OpenState) OnEnter(position Position) error {
	entryDateTime := position.GetEntryOrder().GetExecutionInfo().Datetime
	position.SetEntryDateTime(entryDateTime)
	return nil
}

func (o *OpenState) IsOpen(position Position) bool {
	return true
}

func (o *OpenState) Exit(pos Position, stopPrice, limitPrice float64, goodTillCanceled bool) error {
	if pos.GetShares() == 0 {
		return fmt.Errorf("no shares")
	}
	if pos.ExitActive() {
		return fmt.Errorf("exit oder is active and it should be cancelled first")
	}
	if pos.EntryActive() {
		pos.GetStrategy().GetBroker().CancelOrder(pos.GetEntryOrder())
	}
	if pos2, ok := pos.(*position); ok {
		pos2.submitExitOrder(stopPrice, limitPrice, goodTillCanceled)
	} else {
		return fmt.Errorf("failed to submit exit order")
	}

	return nil
}

type ClosedState struct {
}

func (c *ClosedState) CanSubmitOrder(position Position, order common.Order) bool {
	return false
}

func (c *ClosedState) OnOrderEvent(position Position, orderEvent *common.OrderEvent) error {
	return nil
}

func (c *ClosedState) OnEnter(position Position) error {
	if position.ExitFilled() {
		exitDateTime := position.GetExitOrder().GetExecutionInfo().Datetime
		position.SetExitDateTime(exitDateTime)
	}
	if position.GetShares() == 0 {
		return fmt.Errorf("no shares")
	}
	position.GetStrategy().UnregisterPosition(position)
	return nil
}

func (c *ClosedState) IsOpen(position Position) bool {
	return false
}

func (c *ClosedState) Exit(position Position, stopPrice, limitPrice float64, goodTillCanceled bool) error {
	return nil
}
