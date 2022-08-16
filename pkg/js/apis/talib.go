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
	TALib   *talib.TALib
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
	t.TALib = ta
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

		if numArgs-1 != len(call.ArgumentList) {
			logger.Logger.Debug("talib method needs correct number of arguments",
				zap.String("method", name),
				zap.Int("expected", numArgs-1),
				zap.Int("actual", len(call.ArgumentList)))
			return otto.NullValue()
		}

		// convert otto.Value to reflect.Value
		for i := 0; i < numArgs; i++ {
			if i == 0 {
				args[0] = reflect.ValueOf(t.TALib)
				continue
			}
			if call.Argument(i - 1).IsObject() {
				// Object
				obj := call.Argument(i - 1).Object()
				if obj.Class() == "Array" &&
					(method.Type.In(i).Kind() == reflect.Array ||
						method.Type.In(i).Kind() == reflect.Slice) {
					inArgRaw, err := obj.Value().Export()
					if err != nil {
						logger.Logger.Debug("talib method argument is not an array")
						return otto.NullValue()
					}

					switch v := inArgRaw.(type) {
					case []float64:
						args[i] = reflect.ValueOf(v)
					case []float32:
					case []int:
					case []int32:
					case []int64:
						v2 := make([]float64, len(v))
						for i := 0; i < len(v); i++ {
							v2[i] = float64(v[i])
						}
						args[i] = reflect.ValueOf(v2)
					default:
						logger.Logger.Debug("talib method argument unknown type")
						return otto.NullValue()
					}
				} else {
					logger.Logger.Debug("talib method argument is not an array")
					return otto.NullValue()
				}
			} else if call.Argument(i - 1).IsString() {
				// String
				logger.Logger.Debug("talib method argument is a string, not supported")
				return otto.NullValue()
			} else if call.Argument(i - 1).IsBoolean() {
				// Boolean
				logger.Logger.Debug("talib method argument is a boolean, not supported")
				return otto.NullValue()
			} else if call.Argument(i - 1).IsNumber() {
				// Number
				inArgInt, err := call.Argument(i - 1).ToInteger()
				if err != nil {
					logger.Logger.Debug("talib method argument is not a number")
					return otto.NullValue()
				}
				args[i] = reflect.ValueOf(int(inArgInt))
			} else {
				logger.Logger.Debug("talib method argument is not supported",
					zap.String("method", name),
					zap.Int("index", i),
					zap.Any("value", call.Argument(i-1)))
				return otto.NullValue()
			}
		}

		rtn := method.Func.Call(args)

		rtnVal := []interface{}{}
		for _, v := range rtn {
			rtnVal = append(rtnVal, v.Interface())
		}

		if val, err := t.VM.ToValue(rtnVal); err != nil {
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
		// logger.Logger.Debug("registering talib method", zap.String("method", k))
		t.registerSingleMethod(obj, k, v)
	}
}
