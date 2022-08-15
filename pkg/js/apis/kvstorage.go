package apis

import (
	"goalgotrade/pkg/logger"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

type KVObject struct {
	VM     *otto.Otto
	DBPath string
	DB     *badger.DB
}

var CleanUpDuration = time.Second * 30

func NewKVObject(vm *otto.Otto, dbFilePath string) (*KVObject, error) {
	kv := &KVObject{
		VM:     vm,
		DBPath: dbFilePath,
	}

	if dbFilePath != "" {
		db, err := badger.Open(badger.DefaultOptions(dbFilePath))
		if err != nil {
			logger.Logger.Fatal("failed to open badger db file", zap.Error(err))
			return nil, err
		}
		kv.DB = db
	} else {
		db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
		if err != nil {
			logger.Logger.Fatal("failed to open in-memory badger db", zap.Error(err))
			return nil, err
		}
		kv.DB = db
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
		err := kv.DB.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}

func (kv *KVObject) DBLoadState(key []byte) ([]byte, error) {
	var data []byte
	err := kv.DB.View(func(txn *badger.Txn) error {
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
	return kv.DB.Update(func(txn *badger.Txn) error {
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
