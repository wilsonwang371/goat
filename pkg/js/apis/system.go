package apis

import (
	"fmt"
	"time"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/dop251/goja"
)

type StartCallback func() error

type SysObject struct {
	cfg *config.Config
	VM  *goja.Runtime
	Cb  StartCallback
}

func NewSysObject(cfg *config.Config, vm *goja.Runtime, startCb StartCallback) (*SysObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	sys := &SysObject{
		cfg: cfg,
		VM:  vm,
		Cb:  startCb,
	}

	sysObj := sys.VM.NewObject()
	sysObj.Set("start", sys.StartCmd)
	sysObj.Set("now", sys.TimeCmd)
	sys.VM.Set("system", sysObj)

	consoleObj := sys.VM.NewObject()
	consoleObj.Set("log", sys.LogCmd)
	sys.VM.Set("console", consoleObj)

	return sys, nil
}

func (sys *SysObject) LogCmd(call goja.FunctionCall) goja.Value {
	for _, arg := range call.Arguments {
		fmt.Print(arg.String())
	}
	fmt.Println()
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
