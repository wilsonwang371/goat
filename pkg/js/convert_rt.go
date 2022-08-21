package js

import (
	"goat/pkg/config"
	"goat/pkg/core"
	"goat/pkg/js/apis"

	"github.com/robertkrimen/otto"
)

type ConvertRuntime interface {
	Compile(source string) (*otto.Script, error)
	Execute(script *otto.Script) (otto.Value, error)
	Convert(data map[string]string) (core.Bar, error)
}

type convertRt struct {
	cfg *config.Config
	vm  *otto.Otto
}

// Compile implements ConvertRuntime
func (*convertRt) Compile(source string) (*otto.Script, error) {
	panic("unimplemented")
}

// Convert implements ConvertRuntime
func (*convertRt) Convert(data map[string]string) (core.Bar, error) {
	panic("unimplemented")
}

// Execute implements ConvertRuntime
func (*convertRt) Execute(script *otto.Script) (otto.Value, error) {
	panic("unimplemented")
}

func NewDBConvertRuntime(cfg *config.Config, cb apis.StartCallback) ConvertRuntime {
	// TODO: implement me
	res := &convertRt{}
	return res
}
