package apis

import (
	"fmt"
	"sync"
	"time"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/dop251/goja"
)

type StartCallback func() error

type SysObject struct {
	cfg *config.Config
	VM  *goja.Runtime
	Mu  *sync.Mutex
	Cb  StartCallback
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
	sys.VM.Set("system", sysObj)

	consoleObj := sys.VM.NewObject()
	consoleObj.Set("log", sys.LogCmd)
	sys.VM.Set("console", consoleObj)

	sys.VM.Set("setInterval", sys.SetIntervalCmd)

	return sys, nil
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
				cb(goja.Undefined())
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
