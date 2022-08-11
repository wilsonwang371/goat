package js

import (
	"strings"

	"goalgotrade/pkg/logger"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
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
		return fn(call)
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

	res.setupStrategyAPIs()

	return res
}

func (r *runtime) storeState(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 2 {
		logger.Logger.Debug("storeState needs 2 arguments")
		return otto.FalseValue()
	}
	key, err := call.Argument(0).ToString()
	if err != nil {
		logger.Logger.Debug("failed to convert key to string", zap.Error(err))
		return otto.FalseValue()
	}
	data, err := call.Argument(1).ToString()
	if err != nil {
		logger.Logger.Debug("failed to convert data to string", zap.Error(err))
		return otto.FalseValue()
	}
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
