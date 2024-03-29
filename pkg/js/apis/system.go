package apis

import (
	"fmt"
	"sync"
	"time"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"go.uber.org/zap"
)

type StartCallback func() error

type SysObject struct {
	cfg            *config.Config
	VM             *goja.Runtime
	Mu             *sync.Mutex
	Cb             StartCallback
	StrategyStatus string
}

func NewSysObject(cfg *config.Config, vm *goja.Runtime, runMu *sync.Mutex, startCb StartCallback) (*SysObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	sys := &SysObject{
		cfg: cfg,
		VM:  vm,
		Mu:  runMu,
		Cb:  startCb,
	}

	sysObj := sys.VM.NewObject()
	sysObj.Set("start", sys.StartCmd)
	sysObj.Set("now", sys.TimeCmd)
	sysObj.Set("strftime", sys.StrftimeCmd)
	sysObj.Set("reportStatus", sys.ReportStatusCmd)
	sys.VM.Set("system", sysObj)

	consoleObj := sys.VM.NewObject()
	consoleObj.Set("log", sys.LogCmd)
	sys.VM.Set("console", consoleObj)

	sys.VM.Set("setInterval", sys.SetIntervalCmd)
	sys.registerRequire()

	return sys, nil
}

func (sys *SysObject) registerRequire() {
	registry := require.Registry{}
	registry.Enable(sys.VM)
}

func (sys *SysObject) ReportStatusCmd(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 1 {
		logger.Logger.Debug("reportStatusCmd needs 1 argument")
		return sys.VM.ToValue(false)
	}

	sys.StrategyStatus = call.Argument(0).String()

	return sys.VM.ToValue(true)
}

func (sys *SysObject) SetIntervalCmd(call goja.FunctionCall) goja.Value {
	logger.Logger.Debug("setIntervalCmd")
	if len(call.Arguments) != 2 {
		logger.Logger.Debug("setIntervalCmd needs 2 argument")
		return sys.VM.ToValue(false)
	}
	if cb, ok := goja.AssertFunction(call.Argument(0)); ok {
		interval := call.Argument(1).ToInteger()
		go func(cb goja.Callable, interval int64, mu *sync.Mutex) {
			for {
				time.Sleep(time.Duration(interval) * time.Millisecond)
				mu.Lock()
				if _, err := cb(goja.Undefined()); err != nil {
					logger.Logger.Error("setIntervalCmd callback error", zap.Error(err))
				}
				mu.Unlock()
			}
		}(cb, interval, sys.Mu)
	}
	return sys.VM.ToValue(true)
}

func (sys *SysObject) StrftimeCmd(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		logger.Logger.Debug("strftimeCmd needs 2 argument")
		return goja.Null()
	}

	format := call.Argument(0).String()
	tm := time.Unix(call.Argument(1).ToInteger(), 0)

	return sys.VM.ToValue(tm.Format(format))
}

func (sys *SysObject) LogCmd(call goja.FunctionCall) goja.Value {
	res := ""
	for _, arg := range call.Arguments {
		res = res + arg.String()
	}
	logger.Logger.Info(res)
	return goja.Undefined()
}

func (sys *SysObject) StartCmd(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 0 {
		logger.Logger.Debug("startCmd needs 0 argument")
		return sys.VM.ToValue(false)
	}

	if sys.Cb == nil {
		logger.Logger.Debug("startCmd callback is nil")
		return sys.VM.ToValue(false)
	}

	if err := sys.Cb(); err != nil {
		return sys.VM.ToValue(false)
	}

	return sys.VM.ToValue(true)
}

func (sys *SysObject) TimeCmd(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 0 {
		logger.Logger.Debug("startCmd needs 0 argument")
		return sys.VM.ToValue(false)
	}

	tm := time.Now().Unix()

	return sys.VM.ToValue(tm)
}
