package js

import (
	"runtime/debug"
	"strings"

	"goat/pkg/config"
	"goat/pkg/core"
	"goat/pkg/js/apis"
	"goat/pkg/logger"

	"github.com/dop251/goja"
	"go.uber.org/zap"

	talib "github.com/wilsonwang371/go-talib"
)

var supportedEvents []string = []string{
	"onbars",
	"onstart",
	"onfinish",
	"onidle",
}

type RuntimeFunc func(call goja.FunctionCall) goja.Value

type StrategyRuntime interface {
	Compile(source string) (*goja.Program, error)
	Execute(script *goja.Program) (goja.Value, error)
	RegisterHostCall(name string, fn RuntimeFunc) error
	NotifyEvent(eventName string, args ...interface{}) error
}

type strategyRuntime struct {
	cfg            *config.Config
	vm             *goja.Runtime
	kvApi          *apis.KVObject
	tlApi          *apis.TALib
	sysApi         *apis.SysObject
	alertApi       *apis.AlertObject
	feedApi        *apis.FeedObject
	eventListeners map[string]goja.Value
	apiHandlers    map[string]RuntimeFunc
	talib          *talib.TALib
}

// NotifyEvent implements StrategyRuntime
func (r *strategyRuntime) NotifyEvent(eventName string, args ...interface{}) error {
	if handler, ok := r.eventListeners[strings.ToLower(eventName)]; ok {
		var handlerFunc func(...interface{}) goja.Value
		if err := r.vm.ExportTo(handler, &handlerFunc); err != nil {
			return err
		} else {
			handlerFunc(args...)
		}
	}
	return nil
}

// RegisterHostCall implements StrategyRuntime
func (r *strategyRuntime) RegisterHostCall(name string, fn RuntimeFunc) error {
	return r.vm.Set(name, func(call goja.FunctionCall) goja.Value {
		defer func() {
			if r := recover(); r != nil {
				logger.Logger.Debug("strategyRuntime panic", zap.Any("panic", r))
				logger.Logger.Debug(string(debug.Stack()))
			}
		}()
		rtn := goja.Null()
		rtn = fn(call)
		return rtn
	})
}

// Execute implements StrategyRuntime
func (r *strategyRuntime) Execute(script *goja.Program) (goja.Value, error) {
	return r.vm.RunProgram(script)
}

// Compile implements StrategyRuntime
func (r *strategyRuntime) Compile(source string) (*goja.Program, error) {
	compiled, err := goja.Compile("", source, true)
	if err != nil {
		return nil, err
	}
	return compiled, nil
}

func NewStrategyRuntime(cfg *config.Config, feed core.DataFeed, cb apis.StartCallback) StrategyRuntime {
	var err error

	res := &strategyRuntime{
		cfg:            cfg,
		vm:             goja.New(),
		apiHandlers:    make(map[string]RuntimeFunc),
		eventListeners: make(map[string]goja.Value),
		talib:          talib.NewTALib(),
	}

	logger.Logger.Debug("using kvdb file.", zap.String("kvdb", cfg.KVDB))

	res.kvApi, err = apis.NewKVObject(cfg, res.vm, cfg.KVDB)
	if err != nil {
		logger.Logger.Error("failed to create kv object", zap.Error(err))
		panic(err)
	}
	res.tlApi, err = apis.NewTALibObject(cfg, res.vm)
	if err != nil {
		logger.Logger.Error("failed to create talib object", zap.Error(err))
		panic(err)
	}
	res.sysApi, err = apis.NewSysObject(cfg, res.vm, cb)
	if err != nil {
		logger.Logger.Error("failed to create sys object", zap.Error(err))
		panic(err)
	}
	res.alertApi, err = apis.NewAlertObject(cfg, res.vm)
	if err != nil {
		logger.Logger.Error("failed to create alert object", zap.Error(err))
		panic(err)
	}
	res.feedApi, err = apis.NewFeedObject(cfg, res.vm, feed)
	if err != nil {
		logger.Logger.Error("failed to create feed object", zap.Error(err))
		panic(err)
	}

	res.apiHandlers["addEventListener"] = res.addEventListener
	res.setupStrategyAPIs()

	return res
}

func (r *strategyRuntime) addEventListener(call goja.FunctionCall) goja.Value {
	// logger.Logger.Info("addEventListener is called")
	if len(call.Arguments) != 2 {
		logger.Logger.Error("addEventListener needs 2 arguments")
		return r.vm.ToValue(false)
	}
	eventName := call.Argument(0).String()
	handler := call.Argument(1)

	if !contains(supportedEvents, strings.ToLower(eventName)) {
		logger.Logger.Error("unsupported event", zap.String("event", eventName))
		return r.vm.ToValue(false)
	}

	r.eventListeners[strings.ToLower(eventName)] = handler
	return r.vm.ToValue(true)
}

func (r *strategyRuntime) setupStrategyAPIs() {
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
