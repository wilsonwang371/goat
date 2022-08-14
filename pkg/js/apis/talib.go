package apis

import (
	"reflect"

	"goalgotrade/pkg/logger"

	"github.com/robertkrimen/otto"
	"github.com/wilsonwang371/go-talib"
	"go.uber.org/zap"
)

type TALib struct {
	VM      *otto.Otto
	Methods map[string]reflect.Method
}

func NewTALibObject(vm *otto.Otto) (*TALib, error) {
	t := &TALib{
		VM: vm,
	}
	t.populateMethods()
	obj, err := t.VM.Object(`talib = {}`)
	if err != nil {
		return nil, err
	}
	t.registerMethods(obj)

	return t, nil
}

func (t *TALib) populateMethods() {
	ta := talib.NewTALib()
	r := reflect.TypeOf(ta)
	t.Methods = make(map[string]reflect.Method)
	for i := 0; i < r.NumMethod(); i++ {
		if r.Method(i).Name != "" {
			t.Methods[r.Method(i).Name] = r.Method(i)
		}
	}
}

func (t *TALib) registerSingleMethod(obj *otto.Object, name string, method reflect.Method) {
	obj.Set(name, func(call otto.FunctionCall) otto.Value {
		logger.Logger.Debug("calling talib method", zap.String("method", name))
		numArgs := method.Type.NumIn()
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

		rtn := method.Func.Call(args)
		if len(rtn) != 1 {
			logger.Logger.Error("talib method returned more than one value",
				zap.String("method", name))
		}
		if val, err := otto.ToValue(rtn[0].Interface()); err != nil {
			logger.Logger.Error("talib method returned invalid value",
				zap.String("method", name))
			return otto.NullValue()
		} else {
			return val
		}
	})
}

func (t *TALib) registerMethods(obj *otto.Object) {
	for k, v := range t.Methods {
		logger.Logger.Debug("registering talib method", zap.String("method", k))
		t.registerSingleMethod(obj, k, v)
	}
}
