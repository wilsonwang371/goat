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

func (t *TALib) registerSingleMethod(obj *goja.Object, name string, method reflect.Method) {
	obj.Set(name, func(call goja.FunctionCall) goja.Value {
		logger.Logger.Debug("calling talib method", zap.String("method", name))
		numArgs := method.Type.NumIn()
		args := make([]reflect.Value, numArgs)

		if numArgs-1 != len(call.Arguments) {
			logger.Logger.Debug("talib method needs correct number of arguments",
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

			{
				var val []float64
				if t.VM.ExportTo(call.Argument(i-1), &val) != nil {
					args[i] = reflect.ValueOf(val)
					continue
				}
			}

			{
				var val []float32
				if t.VM.ExportTo(call.Argument(i-1), &val) != nil {
					v2 := make([]float64, len(val))
					for i := 0; i < len(val); i++ {
						v2[i] = float64(val[i])
					}
					args[i] = reflect.ValueOf(v2)
					continue
				}
			}

			{
				var val []int32
				if t.VM.ExportTo(call.Argument(i-1), &val) != nil {
					v2 := make([]float64, len(val))
					for i := 0; i < len(val); i++ {
						v2[i] = float64(val[i])
					}
					args[i] = reflect.ValueOf(v2)
					continue
				}
			}

			{
				var val []int64
				if t.VM.ExportTo(call.Argument(i-1), &val) != nil {
					v2 := make([]float64, len(val))
					for i := 0; i < len(val); i++ {
						v2[i] = float64(val[i])
					}
					args[i] = reflect.ValueOf(v2)
					continue
				}
			}

			{
				var val bool
				if t.VM.ExportTo(call.Argument(i-1), &val) != nil {
					args[i] = reflect.ValueOf(val)
					continue
				}
			}

			{
				var val string
				if t.VM.ExportTo(call.Argument(i-1), &val) != nil {
					args[i] = reflect.ValueOf(val)
					continue
				}
			}

			{
				var val int64
				if t.VM.ExportTo(call.Argument(i-1), &val) != nil {
					args[i] = reflect.ValueOf(val)
					continue
				}
			}

			logger.Logger.Debug("talib method argument is not supported",
				zap.String("method", name),
				zap.Int("index", i),
				zap.Any("value", call.Argument(i-1)))
			return goja.Null()
		}

		rtn := method.Func.Call(args)

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
