package js

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"

	"goalgotrade/pkg/logger"

	badger "github.com/dgraph-io/badger/v3"
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
	db             *badger.DB
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

func NewRuntime(dbFilePath string) Runtime {
	res := &runtime{
		vm:             otto.New(),
		apiHandlers:    make(map[string]RuntimeFunc),
		eventListeners: make(map[string]otto.Value),
		talib:          talib.NewTALib(),
	}

	if dbFilePath != "" {
		db, err := badger.Open(badger.DefaultOptions(dbFilePath))
		if err != nil {
			logger.Logger.Fatal("failed to open badger db file", zap.Error(err))
		}
		res.db = db
	} else {
		db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
		if err != nil {
			logger.Logger.Fatal("failed to open in-memory badger db", zap.Error(err))
		}
		res.db = db
	}

	res.apiHandlers["addEventListener"] = res.addEventListener
	res.apiHandlers["store"] = res.storeState
	res.apiHandlers["load"] = res.loadState
	res.addTALibMethods()
	res.setupStrategyAPIs()

	return res
}

func (r *runtime) addTALibMethods() {
	for _, talibMethod := range AllTALibMethods() {
		name := fmt.Sprintf("%s_%s", "talib", talibMethod.Name)
		logger.Logger.Debug("registering talib method", zap.String("method", name))
		r.apiHandlers[name] = func(call otto.FunctionCall) otto.Value {
			logger.Logger.Debug("calling talib method", zap.String("method", name))
			numArgs := talibMethod.Type.NumIn()
			args := make([]reflect.Value, numArgs)

			if numArgs != len(call.ArgumentList) {
				logger.Logger.Debug("talib method needs correct number of arguments",
					zap.String("method", name),
					zap.Int("expected", numArgs),
					zap.Int("actual", len(call.ArgumentList)))
				return otto.NullValue()
			}

			// convert otto.Value to reflect.Value
			for i := 0; i < numArgs; i++ {
				obj := call.Argument(i).Object()
				if obj == nil {
					logger.Logger.Debug("talib method argument is not an object",
						zap.String("method", name))
					return otto.NullValue()
				}
				if newVal, err := obj.Value().Export(); err != nil {
					logger.Logger.Debug("failed to convert otto.Value to reflect.Value",
						zap.Error(err))
					return otto.NullValue()
				} else {
					args[i] = reflect.ValueOf(newVal)
				}
			}

			rtn := talibMethod.Func.Call(args)
			if len(rtn) != 1 {
				logger.Logger.Error("talib method returned more than one value",
					zap.String("method", talibMethod.Name))
			}
			if val, err := otto.ToValue(rtn[0].Interface()); err != nil {
				logger.Logger.Error("talib method returned invalid value",
					zap.String("method", talibMethod.Name))
				return otto.NullValue()
			} else {
				return val
			}
		}
	}
}

func (r *runtime) storeState(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 2 {
		logger.Logger.Debug("storeState needs 2 arguments")
		return otto.FalseValue()
	}
	for i := 0; i < len(call.ArgumentList); i++ {
		if !call.ArgumentList[i].IsString() {
			logger.Logger.Debug("storeState needs string arguments")
			return otto.FalseValue()
		}
	}
	key := call.Argument(0).String()
	data := call.Argument(1).String()
	if err := r.dbStoreState([]byte(key), []byte(data)); err != nil {
		logger.Logger.Debug("failed to store state", zap.Error(err))
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

func (r *runtime) loadState(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 1 {
		logger.Logger.Debug("loadState needs 1 argument")
		return otto.NullValue()
	}
	for i := 0; i < len(call.ArgumentList); i++ {
		if !call.ArgumentList[i].IsString() {
			logger.Logger.Debug("loadState needs string arguments")
			return otto.NullValue()
		}
	}
	key := call.Argument(0).String()
	data, err := r.dbLoadState([]byte(key))
	if err != nil {
		logger.Logger.Debug("failed to load state", zap.Error(err))
		return otto.NullValue()
	}
	if val, err := otto.ToValue(string(data)); err != nil {
		logger.Logger.Debug("failed to convert data to otto.Value", zap.Error(err))
		return otto.NullValue()
	} else {
		return val
	}
}

func (r *runtime) dbLoadState(key []byte) ([]byte, error) {
	var data []byte
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
		return nil
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (r *runtime) dbStoreState(key []byte, data []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
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
