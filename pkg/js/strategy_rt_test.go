package js

import (
	"context"
	"os"
	"testing"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/dop251/goja"
	"go.uber.org/zap"
)

func TestRuntimeSimple(t *testing.T) {
	os.RemoveAll("default.kvdb")
	defer os.RemoveAll("default.kvdb")
	cfg := &config.Config{
		KVDB: "default.kvdb",
	}
	rt := NewStrategyRuntime(context.TODO(), cfg, nil, nil)
	err := rt.RegisterHostCall("test_print", func(call goja.FunctionCall) goja.Value {
		logger.Logger.Info("test_print is called")
		return goja.Null()
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
	os.RemoveAll("default.kvdb")
	defer os.RemoveAll("default.kvdb")
	cfg := &config.Config{
		KVDB: "default.kvdb",
	}
	rt := NewStrategyRuntime(context.TODO(), cfg, nil, nil)
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
	os.RemoveAll("default.kvdb")
	defer os.RemoveAll("default.kvdb")
	cfg := &config.Config{
		KVDB: "default.kvdb",
	}
	rt := NewStrategyRuntime(context.TODO(), cfg, nil, nil)
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
	os.RemoveAll("default.kvdb")
	defer os.RemoveAll("default.kvdb")
	cfg := &config.Config{
		KVDB: "default.kvdb",
	}
	rt := NewStrategyRuntime(context.TODO(), cfg, nil, nil)
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

func TestRuntimeRequire(t *testing.T) {
	os.RemoveAll("default.kvdb")
	defer os.RemoveAll("default.kvdb")
	cfg := &config.Config{
		KVDB: "default.kvdb",
	}
	rt := NewStrategyRuntime(context.TODO(), cfg, nil, nil)
	script, err := rt.Compile(`
	var m = require("../../samples/misc/require-test.js");
	m.test();
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
