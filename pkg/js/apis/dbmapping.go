package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"

	otto "github.com/dop251/goja"
	"go.uber.org/zap"
)

type DBMappingObject struct {
	cfg      *config.Config
	VM       *otto.Runtime
	Mappings map[string]interface{}
}

func NewDBMappingObject(cfg *config.Config, vm *otto.Runtime) (*DBMappingObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	db := &DBMappingObject{
		cfg:      cfg,
		VM:       vm,
		Mappings: nil,
	}

	dbObj := db.VM.NewObject()
	dbObj.Set("set_mappings", db.SetDBMappingCmd)
	db.VM.Set("dbconvert", dbObj)

	return db, nil
}

func (db *DBMappingObject) SetDBMappingCmd(call otto.FunctionCall) otto.Value {
	if len(call.Arguments) != 1 {
		logger.Logger.Debug("set_mappings needs 1 argument")
		return db.VM.ToValue(false)
	}

	if call.Argument(0).ToObject(db.VM) != nil {
		rawObj := call.Argument(0).Export()
		if rawObj == nil {
			logger.Logger.Debug("set_mappings argument is nil")
			return db.VM.ToValue(false)
		}
		logger.Logger.Debug("set_mappings", zap.Any("obj", rawObj))

		if obj, ok := rawObj.(map[string]interface{}); ok {
			db.Mappings = obj
		}
	} else {
		logger.Logger.Debug("invalid argument")
	}

	return db.VM.ToValue(true)
}
