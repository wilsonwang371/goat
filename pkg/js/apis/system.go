package apis

import (
	"goat/pkg/logger"

	"github.com/robertkrimen/otto"
)

type StartCallback func() error

type SysObject struct {
	VM *otto.Otto
	Cb StartCallback
}

func NewSysObject(vm *otto.Otto, startCb StartCallback) (*SysObject, error) {
	sys := &SysObject{
		VM: vm,
		Cb: startCb,
	}

	sysObj, err := sys.VM.Object(`system = {}`)
	if err != nil {
		return nil, err
	}
	sysObj.Set("start", sys.StartCmd)

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
