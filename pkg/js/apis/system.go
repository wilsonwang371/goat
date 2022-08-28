package apis

import (
	"fmt"
	"time"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/robertkrimen/otto"
)

type StartCallback func() error

type SysObject struct {
	cfg *config.Config
	VM  *otto.Otto
	Cb  StartCallback
}

func NewSysObject(cfg *config.Config, vm *otto.Otto, startCb StartCallback) (*SysObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	sys := &SysObject{
		cfg: cfg,
		VM:  vm,
		Cb:  startCb,
	}

	sysObj, err := sys.VM.Object(`system = {}`)
	if err != nil {
		return nil, err
	}
	sysObj.Set("start", sys.StartCmd)
	sysObj.Set("now", sys.TimeCmd)

	return sys, nil
}

func (sys *SysObject) StartCmd(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 0 {
		logger.Logger.Debug("startCmd needs 0 argument")
		return otto.FalseValue()
	}

	if sys.Cb == nil {
		logger.Logger.Debug("startCmd callback is nil")
		return otto.TrueValue()
	}

	if err := sys.Cb(); err != nil {
		return otto.FalseValue()
	}

	return otto.TrueValue()
}

func (sys *SysObject) TimeCmd(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 0 {
		logger.Logger.Debug("startCmd needs 0 argument")
		return otto.FalseValue()
	}

	tm := time.Now().Unix()

	if val, err := sys.VM.ToValue(tm); err != nil {
		return otto.NaNValue()
	} else {
		return val
	}
}
