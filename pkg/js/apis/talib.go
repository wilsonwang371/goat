package apis

import (
	"fmt"
	"reflect"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/dop251/goja"
	"github.com/wilsonwang371/go-talib"
	"go.uber.org/zap"
)

type TALib struct {
	cfg     *config.Config
	VM      *goja.Runtime
	Methods map[string]reflect.Method
	TALib   *talib.TALib
}

func NewTALibObject(cfg *config.Config, vm *goja.Runtime) (*TALib, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	t := &TALib{
		cfg: cfg,
		VM:  vm,
	}
	t.populateMethods()
	obj := t.VM.NewObject()
	t.registerMethods(obj)
	if err := vm.Set("talib", obj); err != nil {
		logger.Logger.Error("failed to register talib object", zap.Error(err))
		return nil, err
	}

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

func isInterfaceArrayFloatArray(arr []interface{}) bool {
	floatAmount := 0
	for _, v := range arr {
		switch v.(type) {
		case float32:
		case float64:
			floatAmount += 1
		}
	}
	if floatAmount != 0 {
		return true
	}
	return false
}

func convertInterfaceArrayToIntArray(arr []interface{}) []int64 {
	var ret []int64
	for _, v := range arr {
		switch v.(type) {
		case float32:
			ret = append(ret, int64(v.(float32)))
		case float64:
			ret = append(ret, int64(v.(float64)))
		case int:
			ret = append(ret, int64(v.(int)))
		case int32:
			ret = append(ret, int64(v.(int32)))
		case int64:
			ret = append(ret, v.(int64))
		default:
			logger.Logger.Error("unsupported type", zap.Any("type", v))
			return nil
		}
	}
	return ret
}

func convertInterfaceArrayToFloatArray(arr []interface{}) []float64 {
	var ret []float64
	for _, v := range arr {
		switch v.(type) {
		case float32:
			ret = append(ret, float64(v.(float32)))
		case float64:
			ret = append(ret, v.(float64))
		case int:
			ret = append(ret, float64(v.(int)))
		case int32:
			ret = append(ret, float64(v.(int32)))
		case int64:
			ret = append(ret, float64(v.(int64)))
		default:
			logger.Logger.Error("unsupported type", zap.Any("type", v))
			return nil
		}
	}
	return ret
}

func (t *TALib) registerSingleMethod(obj *goja.Object, name string, method reflect.Method) {
	obj.Set(name, func(call goja.FunctionCall) goja.Value {
		logger.Logger.Debug("calling talib method", zap.String("method", name))
		numArgs := method.Type.NumIn()
		args := make([]reflect.Value, numArgs)

		if numArgs-1 != len(call.Arguments) {
			logger.Logger.Info("talib method needs correct number of arguments",
				zap.String("method", name),
				zap.Int("expected", numArgs-1),
				zap.Int("actual", len(call.Arguments)))
			return goja.Null()
		}

		// convert goja.Value to reflect.Value
		for i := 0; i < numArgs; i++ {
			if i == 0 {
				args[0] = reflect.ValueOf(t.TALib)
				continue
			}

			inArgRaw := call.Arguments[i-1].Export()
			switch v := inArgRaw.(type) {
			case []interface{}:
				if len(v) == 0 {
					args[i] = reflect.ValueOf([]float64{})
				}
				if isInterfaceArrayFloatArray(v) {
					args[i] = reflect.ValueOf(convertInterfaceArrayToFloatArray(v))
				} else {
					args[i] = reflect.ValueOf(convertInterfaceArrayToIntArray(v))
				}
			case int:
				if method.Type.In(i).ConvertibleTo(reflect.TypeOf(v)) {
					args[i] = reflect.ValueOf(v).Convert(method.Type.In(i))
				} else {
					logger.Logger.Error("talib method argument type mismatch",
						zap.String("method", name),
						zap.Int("index", i),
						zap.String("expected", method.Type.In(i).String()),
						zap.String("actual", reflect.TypeOf(v).String()))
					return goja.Null()
				}
			case int64:
				if method.Type.In(i).ConvertibleTo(reflect.TypeOf(v)) {
					args[i] = reflect.ValueOf(v).Convert(method.Type.In(i))
				} else {
					logger.Logger.Error("talib method argument type mismatch",
						zap.String("method", name),
						zap.Int("index", i),
						zap.String("expected", method.Type.In(i).String()),
						zap.String("actual", reflect.TypeOf(v).String()))
					return goja.Null()
				}
			case int32:
				if method.Type.In(i).ConvertibleTo(reflect.TypeOf(v)) {
					args[i] = reflect.ValueOf(v).Convert(method.Type.In(i))
				} else {
					logger.Logger.Error("talib method argument type mismatch",
						zap.String("method", name),
						zap.Int("index", i),
						zap.String("expected", method.Type.In(i).String()),
						zap.String("actual", reflect.TypeOf(v).String()))
					return goja.Null()
				}
			case string:
				if method.Type.In(i).ConvertibleTo(reflect.TypeOf(v)) {
					args[i] = reflect.ValueOf(v).Convert(method.Type.In(i))
				} else {
					logger.Logger.Error("talib method argument type mismatch",
						zap.String("method", name),
						zap.Int("index", i),
						zap.String("expected", method.Type.In(i).String()),
						zap.String("actual", reflect.TypeOf(v).String()))
					return goja.Null()
				}
			case bool:
				if method.Type.In(i).ConvertibleTo(reflect.TypeOf(v)) {
					args[i] = reflect.ValueOf(v).Convert(method.Type.In(i))
				} else {
					logger.Logger.Error("talib method argument type mismatch",
						zap.String("method", name),
						zap.Int("index", i),
						zap.String("expected", method.Type.In(i).String()),
						zap.String("actual", reflect.TypeOf(v).String()))
					return goja.Null()
				}
			default:
				logger.Logger.Info("talib method argument unknown type", zap.String("method", name), zap.Any("type", reflect.TypeOf(v)))
				return goja.Null()
			}
		}

		rtn := method.Func.Call(args)

		if len(rtn) == 1 {
			return t.VM.ToValue(rtn[0].Interface())
		}

		rtnVal := []interface{}{}
		for _, v := range rtn {
			rtnVal = append(rtnVal, v.Interface())
		}

		return t.VM.ToValue(rtnVal)
	})
}

func (t *TALib) registerMethods(obj *goja.Object) {
	for k, v := range t.Methods {
		// logger.Logger.Debug("registering talib method", zap.String("method", k))
		t.registerSingleMethod(obj, k, v)
	}
}
