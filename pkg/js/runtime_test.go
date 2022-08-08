package js

import (
	"goalgotrade/pkg/logger"
	"testing"

	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

func TestRuntimeSimple(t *testing.T) {
	rt := NewRuntime()
	err := rt.RegisterHostCall("test_print", func(call otto.FunctionCall) otto.Value {
		logger.Logger.Info("test_print is called")
		return otto.NullValue()
	})
	if err != nil {
		t.Error(err)
	}
	script, err := rt.Compile("test_print(1);")
	if err != nil {
		t.Error(err)
	}
	val, err := rt.Execute(script)
	if err != nil {
		t.Error(err)
	}
	logger.Logger.Info("result:", zap.Any("val", val))
}

func TestRuntimeSimple2(t *testing.T) {
	rt := NewRuntime()
	script, err := rt.Compile(`
	addEventListener("onbar", function(e) {
		console.log("onbar", e);
	});
`)
	if err != nil {
		t.Error(err)
	}
	val, err := rt.Execute(script)
	if err != nil {
		t.Error(err)
	}
	logger.Logger.Info("result:", zap.Any("val", val))

	rt.NotifyEvent("onbar", "foo")
}
