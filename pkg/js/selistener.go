package js

import (
	"encoding/json"

	"goat/pkg/metrics"

	"goat/pkg/core"
	"goat/pkg/logger"

	"go.uber.org/zap"
)

func NewJSStrategyEventListener(rt StrategyRuntime) core.StrategyEventListener {
	return &JSStrategyEventListener{
		rt: rt,
	}
}

type JSStrategyEventListener struct {
	rt StrategyRuntime
}

// OnBars implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnBars(bars core.Bars) error {
	jsonData, err := json.Marshal(bars)
	if err != nil {
		logger.Logger.Error("onBars got invalid data", zap.Error(err))
		return err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		logger.Logger.Error("onBars got invalid data to unmarshal", zap.Error(err))
		return err
	}
	metrics.OnBarsCalledCount.Inc()
	return j.rt.NotifyEvent("onbars", data)
}

// OnFinish implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnFinish(args ...interface{}) error {
	return j.rt.NotifyEvent("onfinish", args)
}

// OnIdle implements core.StrategyEventListener
func (j *JSStrategyEventListener) OnIdle() error {
	metrics.OnIdleCalledCount.Inc()
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
