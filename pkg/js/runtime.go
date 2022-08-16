package js

import (
	"runtime/debug"
	"strings"

	"goalgotrade/pkg/js/apis"
	"goalgotrade/pkg/logger"

	"github.com/robertkrimen/otto"
	"go.uber.org/zap"

	talib "github.com/wilsonwang371/go-talib"
)

var supportedEvents []string = []string{
	"onbars",
	"onstart",
	"onfinish",
	"onidle",
}

type RuntimeFunc func(call otto.FunctionCall) otto.Value

type Runtime interface {
	Compile(source string) (*otto.Script, error)
	Execute(script *otto.Script) (otto.Value, error)
	RegisterHostCall(name string, fn RuntimeFunc) error
	NotifyEvent(eventName string, args ...interface{}) error
}

type runtime struct {
	vm             *otto.Otto
	kv             *apis.KVObject
	tl             *apis.TALib
	sys            *apis.SysObject
	eventListeners map[string]otto.Value
	apiHandlers    map[string]RuntimeFunc
	talib          *talib.TALib
}

// NotifyEvent implements Runtime
func (r *runtime) NotifyEvent(eventName string, args ...interface{}) error {
	if handler, ok := r.eventListeners[strings.ToLower(eventName)]; ok {
		if _, err := handler.Call(otto.NullValue(), args...); err != nil {
			logger.Logger.Debug("failed to call handler", zap.Error(err))
			return err
		}
	}
	return nil
}

// RegisterHostCall implements Runtime
func (r *runtime) RegisterHostCall(name string, fn RuntimeFunc) error {
	return r.vm.Set(name, func(call otto.FunctionCall) otto.Value {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Debug("runtime panic", zap.Any("panic", r))
				logger.Logger.Debug(string(debug.Stack()))
			}
		}()
		rtn := otto.NullValue()
		rtn = fn(call)
		return rtn
	})
}

// Execute implements Runtime
func (r *runtime) Execute(script *otto.Script) (otto.Value, error) {
	return r.vm.Run(script)
}

// Compile implements Runtime
func (r *runtime) Compile(source string) (*otto.Script, error) {
	compiled, err := r.vm.Compile("", source)
	if err != nil {
		return nil, err
	}
	return compiled, nil
}

func NewRuntime(dbFilePath string, cb apis.StartCallback) Runtime {
	var err error

	res := &runtime{
		vm:             otto.New(),
		apiHandlers:    make(map[string]RuntimeFunc),
		eventListeners: make(map[string]otto.Value),
		talib:          talib.NewTALib(),
	}

	logger.Logger.Info("using db file.", zap.String("db", dbFilePath))

	res.kv, err = apis.NewKVObject(res.vm, dbFilePath)
	if err != nil {
		logger.Logger.Error("failed to create kv object", zap.Error(err))
		panic(err)
	}
	res.tl, err = apis.NewTALibObject(res.vm)
	if err != nil {
		logger.Logger.Error("failed to create talib object", zap.Error(err))
		panic(err)
	}
	res.sys, err = apis.NewSysObject(res.vm, cb)
	if err != nil {
		logger.Logger.Error("failed to create sys object", zap.Error(err))
		panic(err)
	}

	res.apiHandlers["addEventListener"] = res.addEventListener
	res.setupStrategyAPIs()

	return res
}

func (r *runtime) addEventListener(call otto.FunctionCall) otto.Value {
	// logger.Logger.Info("addEventListener is called")
	if len(call.ArgumentList) != 2 {
		logger.Logger.Error("addEventListener needs 2 arguments")
		return otto.FalseValue()
	}
	eventName := call.Argument(0).String()
	handler := call.Argument(1)

	if !contains(supportedEvents, strings.ToLower(eventName)) {
		logger.Logger.Error("unsupported event", zap.String("event", eventName))
		return otto.FalseValue()
	}

	r.eventListeners[strings.ToLower(eventName)] = handler
	return otto.TrueValue()
}

func (r *runtime) setupStrategyAPIs() {
	for name, fn := range r.apiHandlers {
		if err := r.RegisterHostCall(name, fn); err != nil {
			logger.Logger.Error("failed to register API", zap.Error(err))
		}
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
