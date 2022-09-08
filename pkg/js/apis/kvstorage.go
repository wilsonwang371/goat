package apis

import (
	"context"
	"fmt"
	"time"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/dgraph-io/badger/v3"
	"github.com/dop251/goja"
	"go.uber.org/zap"
)

type KVObject struct {
	ctx      context.Context
	cfg      *config.Config
	VM       *goja.Runtime
	KVDBPath string
	KVDB     *badger.DB
}

var CleanUpDuration = time.Second * 30

func NewKVObject(ctx context.Context, cfg *config.Config, vm *goja.Runtime, kvdbFilePath string) (*KVObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	kv := &KVObject{
		ctx:      ctx,
		cfg:      cfg,
		VM:       vm,
		KVDBPath: kvdbFilePath,
	}

	if kvdbFilePath != "" {
		kvdb, err := badger.Open(badger.DefaultOptions(kvdbFilePath).WithLoggingLevel(badger.ERROR))
		if err != nil {
			logger.Logger.Fatal("failed to open badger kvdb file", zap.Error(err))
			return nil, err
		}
		kv.KVDB = kvdb
	} else {
		kvdb, err := badger.Open(badger.DefaultOptions("").WithInMemory(true).WithLoggingLevel(badger.ERROR))
		if err != nil {
			logger.Logger.Fatal("failed to open in-memory badger kvdb", zap.Error(err))
			return nil, err
		}
		kv.KVDB = kvdb
	}

	kvObj := kv.VM.NewObject()
	kvObj.Set("save", kv.SaveState)
	kvObj.Set("load", kv.LoadState)

	kv.VM.Set("kvstorage", kvObj)

	go kv.cleanup()

	return kv, nil
}

func (kv *KVObject) cleanup() {
	ticker := time.NewTicker(CleanUpDuration)
	defer ticker.Stop()
	select {
	case <-ticker.C:
		kv.KVDB.RunValueLogGC(0.5)
	case <-kv.ctx.Done():
		kv.KVDB.Close()
		return
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

func (kv *KVObject) SaveState(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 2 {
		logger.Logger.Debug("saveState needs 2 arguments")
		return kv.VM.ToValue(false)
	}
	key := call.Argument(0).String()
	data := call.Argument(1).String()
	if err := kv.DBSaveState([]byte(key), []byte(data)); err != nil {
		logger.Logger.Debug("failed to save state", zap.Error(err))
		return kv.VM.ToValue(false)
	}
	return kv.VM.ToValue(true)
}

func (kv *KVObject) LoadState(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) != 1 {
		logger.Logger.Debug("loadState needs 1 argument")
		return goja.Null()
	}
	key := call.Argument(0).String()
	data, err := kv.DBLoadState([]byte(key))
	if err != nil {
		logger.Logger.Debug("failed to load state", zap.Error(err))
		return goja.Null()
	}
	return kv.VM.ToValue(string(data))
}
