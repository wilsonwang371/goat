package strategy

import (
	"fmt"
	"goalgotrade/common"
	lg "goalgotrade/logger"
	"time"

	"go.uber.org/zap"
)

type enterExitOrder interface {
	GetEntryOrder() common.Order
	GetExitOrder() common.Order
	EntryActive() bool
	EntryFilled() bool
	ExitActive() bool
	ExitFilled() bool
	SetEntryDateTime(dateTime time.Time)
	SetExitDateTime(dateTime time.Time)
}

type Position interface {
	enterExitOrder
	OnOrderEvent(orderEvent *common.OrderEvent) error
	IsOpen() bool
	BuildExitOrder(stopPrice, limitPrice float64) common.Order
	GetShares() int
	GetStrategy() Strategy
	SwitchState(newState PositionState)
	GetActiveOrders() []common.Order
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
	PositionStateWaitingEntryState PositionStateType = iota
	PositionStateOpenState
	PositionStateClosedState
)

type basePosition struct {
	Self  interface{}
	state PositionState

	entryOrder    common.Order
	entryDateTime *time.Time
	exitOrder     common.Order
	exitDateTime  *time.Time
	allOrNone     bool
	activeOrders  map[int]common.Order

	strategy Strategy
	shares   int
}

func NewBasePosition(strategy Strategy, entryOrder common.Order, goodTillCanceled, allOrNone bool) *basePosition {
	res := &basePosition{
		strategy:     strategy,
		entryOrder:   entryOrder,
		allOrNone:    allOrNone,
		activeOrders: map[int]common.Order{},
	}
	// TODO: implement me
	res.Self = res
	return res
}

func (p *basePosition) OnOrderEvent(orderEvent *common.OrderEvent) error {
	return nil
}

func (p *basePosition) IsOpen() bool {
	return p.state.IsOpen(p.Self.(Position))
}

func (p *basePosition) EntryActive() bool {
	return p.entryOrder != nil && p.entryOrder.IsActive()
}

func (p *basePosition) ExitActive() bool {
	return p.exitOrder != nil && p.exitOrder.IsActive()
}

func (p *basePosition) EntryFilled() bool {
	return p.exitOrder != nil && p.entryOrder.IsFilled()
}

func (p *basePosition) ExitFilled() bool {
	return p.exitOrder != nil && p.exitOrder.IsFilled()
}

func (p *basePosition) GetEntryOrder() common.Order {
	return p.entryOrder
}

func (p *basePosition) GetExitOrder() common.Order {
	return p.exitOrder
}

func (p *basePosition) SetEntryDateTime(dateTime time.Time) {
	tmpTime := dateTime
	p.entryDateTime = &tmpTime
}

func (p *basePosition) SetExitDateTime(dateTime time.Time) {
	tmpTime := dateTime
	p.exitDateTime = &tmpTime
}

func (p *basePosition) SwitchState(newState PositionState) {
	p.state = newState
	if err := p.state.OnEnter(p.Self.(Position)); err != nil {
		lg.Logger.Warn("switch state failed", zap.Error(err))
	}
}

func (p *basePosition) GetStrategy() Strategy {
	return p.strategy
}

func (p *basePosition) GetShares() int {
	return p.shares
}

func (p *basePosition) submitExitOrder(stopPrice, limitPrice float64, goodTillCanceled bool) error {
	// TODO: implement me
	return nil
}

func (p *basePosition) BuildExitOrder(stopPrice, limitPrice float64) common.Order {
	panic("not implemented")
	return nil
}

func (p *basePosition) GetAge() *time.Duration {
	if p.entryDateTime != nil {
		var res time.Duration
		if p.exitDateTime != nil {
			res = p.exitDateTime.Sub(*p.exitDateTime)
		} else {
			tmp := p.strategy.GetCurrentDateTime()
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

func (p *basePosition) GetActiveOrders() []common.Order {
	var res []common.Order
	for _, order := range p.activeOrders {
		res = append(res, order)
	}
	return res
}

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

type WaitingEntryState struct{}

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
	err := position.GetStrategy().GetBroker().CancelOrder(position.GetEntryOrder())
	if err != nil {
		return err
	}
	return nil
}

type OpenState struct{}

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
		err := pos.GetStrategy().GetBroker().CancelOrder(pos.GetEntryOrder())
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

type ClosedState struct{}

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
	if err := position.GetStrategy().UnregisterPosition(position); err != nil {
		return err
	}
	return nil
}

func (c *ClosedState) IsOpen(position Position) bool {
	return false
}

func (c *ClosedState) Exit(position Position, stopPrice, limitPrice float64, goodTillCanceled bool) error {
	return nil
}

type LongPosition struct {
	basePosition
}

func NewLongPosition(stopPrice, limitPrice float64) Position {
	// TODO: implement me
	return nil
}

type ShortPosition struct {
	basePosition
}

func NewShortPosition(stopPrice, limitPrice float64) Position {
	// TODO: implement me
	return nil
}
