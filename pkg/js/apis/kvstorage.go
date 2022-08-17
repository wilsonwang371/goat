package apis

import (
	"time"

	"goat/pkg/logger"

	"github.com/dgraph-io/badger/v3"
	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

type KVObject struct {
	VM       *otto.Otto
	KVDBPath string
	KVDB     *badger.DB
}

var CleanUpDuration = time.Second * 30

func NewKVObject(vm *otto.Otto, kvdbFilePath string) (*KVObject, error) {
	kv := &KVObject{
		VM:       vm,
		KVDBPath: kvdbFilePath,
	}

	if kvdbFilePath != "" {
		kvdb, err := badger.Open(badger.DefaultOptions(kvdbFilePath))
		if err != nil {
			logger.Logger.Fatal("failed to open badger kvdb file", zap.Error(err))
			return nil, err
		}
		kv.KVDB = kvdb
	} else {
		kvdb, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
		if err != nil {
			logger.Logger.Fatal("failed to open in-memory badger kvdb", zap.Error(err))
			return nil, err
		}
		kv.KVDB = kvdb
	}

	kvObj, err := kv.VM.Object(`kvstorage = {}`)
	if err != nil {
		return nil, err
	}
	kvObj.Set("save", kv.SaveState)
	kvObj.Set("load", kv.LoadState)

	go kv.cleanup()

	return kv, nil
}

func (kv *KVObject) cleanup() {
	ticker := time.NewTicker(CleanUpDuration)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := kv.KVDB.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}

func (kv *KVObject) DBLoadState(key []byte) ([]byte, error) {
	var data []byte
	err := kv.KVDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
		return nil
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (kv *KVObject) DBSaveState(key []byte, data []byte) error {
	return kv.KVDB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

func (kv *KVObject) SaveState(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 2 {
		logger.Logger.Debug("saveState needs 2 arguments")
		return otto.FalseValue()
	}
	for i := 0; i < len(call.ArgumentList); i++ {
		if !call.ArgumentList[i].IsString() {
			logger.Logger.Debug("saveState needs string arguments")
			return otto.FalseValue()
		}
	}
	key := call.Argument(0).String()
	data := call.Argument(1).String()
	if err := kv.DBSaveState([]byte(key), []byte(data)); err != nil {
		logger.Logger.Debug("failed to save state", zap.Error(err))
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

func (kv *KVObject) LoadState(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 1 {
		logger.Logger.Debug("loadState needs 1 argument")
		return otto.NullValue()
	}
	for i := 0; i < len(call.ArgumentList); i++ {
		if !call.ArgumentList[i].IsString() {
			logger.Logger.Debug("loadState needs string arguments")
			return otto.NullValue()
		}
	}
	key := call.Argument(0).String()
	data, err := kv.DBLoadState([]byte(key))
	if err != nil {
		logger.Logger.Debug("failed to load state", zap.Error(err))
		return otto.NullValue()
	}
	if val, err := otto.ToValue(string(data)); err != nil {
		logger.Logger.Debug("failed to convert data to otto.Value", zap.Error(err))
		return otto.NullValue()
	} else {
		return val
	}
}
