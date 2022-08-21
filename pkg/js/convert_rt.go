package js

import (
	"goat/pkg/config"
	"goat/pkg/convert"
	"goat/pkg/js/apis"
	"goat/pkg/logger"

	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

type ConvertRuntime interface {
	Compile(source string) (*otto.Script, error)
	Execute(script *otto.Script) (otto.Value, error)
	Convert(dbsource convert.DBSource) error
}

type convertRt struct {
	cfg     *config.Config
	vm      *otto.Otto
	mapping *apis.DBMappingObject
}

// Compile implements ConvertRuntime
func (*convertRt) Compile(source string) (*otto.Script, error) {
	panic("unimplemented")
}

// Convert implements ConvertRuntime
func (*convertRt) Convert(dbsource convert.DBSource) error {
	panic("unimplemented")
}

// Execute implements ConvertRuntime
func (*convertRt) Execute(script *otto.Script) (otto.Value, error) {
	panic("unimplemented")
}

func NewDBConvertRuntime(cfg *config.Config) ConvertRuntime {
	var err error
	res := &convertRt{
		cfg: cfg,
		vm:  otto.New(),
	}

	res.mapping, err = apis.NewDBMappingObject(cfg, res.vm)
	if err != nil {
		logger.Logger.Error("failed to create db convert mapping object", zap.Error(err))
		panic(err)
	}

	// TODO: implement me

	return res
}
