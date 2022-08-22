package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

type DBMappingObject struct {
	cfg      *config.Config
	VM       *otto.Otto
	Mappings map[string]interface{}
}

func NewDBMappingObject(cfg *config.Config, vm *otto.Otto) (*DBMappingObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	db := &DBMappingObject{
		cfg:      cfg,
		VM:       vm,
		Mappings: nil,
	}

	dbObj, err := db.VM.Object(`dbconvert = {}`)
	if err != nil {
		return nil, err
	}
	dbObj.Set("set_mappings", db.SetDBMappingCmd)

	return db, nil
}

func (db *DBMappingObject) SetDBMappingCmd(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 1 {
		logger.Logger.Debug("set_mappings needs 1 argument")
		return otto.FalseValue()
	}

	if call.Argument(0).IsObject() {
		rawObj, err := call.Argument(0).Export()
		if err != nil {
			logger.Logger.Debug("set_mappings argument is not an object")
			return otto.FalseValue()
		}
		logger.Logger.Debug("set_mappings", zap.Any("obj", rawObj))

		if obj, ok := rawObj.(map[string]interface{}); ok {
			db.Mappings = obj
		}
	} else {
		logger.Logger.Debug("invalid argument")
	}

	return otto.TrueValue()
}
