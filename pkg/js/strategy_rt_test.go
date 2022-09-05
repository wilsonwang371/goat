package js

import (
	"testing"

	"goat/pkg/config"
	"goat/pkg/logger"

	otto "github.com/dop251/goja"
	"go.uber.org/zap"
)

func TestRuntimeSimple(t *testing.T) {
	cfg := &config.Config{
		KVDB: "",
	}
	rt := NewStrategyRuntime(cfg, nil, nil)
	err := rt.RegisterHostCall("test_print", func(call otto.FunctionCall) otto.Value {
		logger.Logger.Info("test_print is called")
		return otto.Null()
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
	cfg := &config.Config{
		KVDB: "",
	}
	rt := NewStrategyRuntime(cfg, nil, nil)
	script, err := rt.Compile(`
	addEventListener("onbars", function(e) {
		console.log("onbars", e);
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

	rt.NotifyEvent("onbars", "foo")
}

func TestRuntimeKV(t *testing.T) {
	cfg := &config.Config{
		KVDB: "",
	}
	rt := NewStrategyRuntime(cfg, nil, nil)
	script, err := rt.Compile(`
	addEventListener("onbars", function(e) {
		kvstorage.save("foo", "bar");
		console.log(kvstorage.load("foo"));
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

	rt.NotifyEvent("onbars", "foo")
}

func TestRuntimeTALibSimple(t *testing.T) {
	cfg := &config.Config{
		KVDB: "",
	}
	rt := NewStrategyRuntime(cfg, nil, nil)
	script, err := rt.Compile(`
	addEventListener("onbars", function(e) {
		var res = talib.Ema([.1,.2,.3,.4,.5,.6,.7,.8], 4);
		console.log("res"+res);
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

	rt.NotifyEvent("onbars", "foo")
}
