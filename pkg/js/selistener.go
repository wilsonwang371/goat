package js

import (
	"encoding/json"

	"goat/pkg/core"
	"goat/pkg/logger"

	"go.uber.org/zap"
)

func NewJSStrategyEventListener(rt Runtime) core.StrategyEventListener {
	return &JSStrategyEventListener{
		rt: rt,
	}
}

type JSStrategyEventListener struct {
	rt Runtime
}

// OnBars implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnBars(bars core.Bars) error {
	jsonStr, err := json.Marshal(bars)
	if err != nil {
		logger.Logger.Error("onBars got invalid data", zap.Error(err))
		return err
	}
	return j.rt.NotifyEvent("onbars", string(jsonStr))
}

// OnFinish implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnFinish(args ...interface{}) error {
	return j.rt.NotifyEvent("onfinish", args)
}

// OnIdle implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnIdle() error {
	return j.rt.NotifyEvent("onidle")
}

// OnOrderEvent implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnOrderEvent(orderEvent core.OrderEvent) error {
	return j.rt.NotifyEvent("onorderevent", orderEvent)
}

// OnOrderUpdated implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnOrderUpdated(order core.Order) error {
	return j.rt.NotifyEvent("onorderupdated", order)
}

// OnStart implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnStart(args ...interface{}) error {
	return j.rt.NotifyEvent("onstart", args)
}
