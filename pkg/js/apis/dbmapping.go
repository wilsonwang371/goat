package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/robertkrimen/otto"
)

type DBMappingObject struct {
	cfg *config.Config
	VM  *otto.Otto
}

func NewDBMappingObject(cfg *config.Config, vm *otto.Otto) (*DBMappingObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	db := &DBMappingObject{
		cfg: cfg,
		VM:  vm,
	}

	dbObj, err := db.VM.Object(`dbconvert = {}`)
	if err != nil {
		return nil, err
	}
	dbObj.Set("set_mappings", db.SetDBMappingCmd)

	return db, nil
}

func (db *DBMappingObject) SetDBMappingCmd(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 0 {
		logger.Logger.Debug("startCmd needs 0 argument")
		return otto.FalseValue()
	}

	// TOOD: implement me

	return otto.TrueValue()
}
