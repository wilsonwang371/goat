package apis

import (
	"context"
	"fmt"
	"time"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/boltdb/bolt"
	"github.com/dop251/goja"
	"go.uber.org/zap"
)

type KVObject struct {
	ctx      context.Context
	cfg      *config.Config
	VM       *goja.Runtime
	KVDBPath string
	KVDB     *bolt.DB
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
		kvdb, err := bolt.Open(kvdbFilePath, 0o600, nil)
		if err != nil {
			logger.Logger.Fatal("failed to open kvdb file", zap.Error(err))
			return nil, err
		}
		err = kvdb.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("default"))
			if err != nil {
				return fmt.Errorf("could not create root bucket: %v", err)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("could not set up buckets, %v", err)
		}
		kv.KVDB = kvdb
	} else {
		return nil, fmt.Errorf("invalid kvdb file path")
	}

	kvObj := kv.VM.NewObject()
	kvObj.Set("save", kv.SaveState)
	kvObj.Set("load", kv.LoadState)

	kv.VM.Set("kvstorage", kvObj)

	return kv, nil
}

func (kv *KVObject) DBLoadState(key []byte) ([]byte, error) {
	var data []byte
	err := kv.KVDB.View(func(txn *bolt.Tx) error {
		item := txn.Bucket([]byte("default")).Get([]byte(key))
		if item == nil {
			return fmt.Errorf("could not find key: %v", key)
		}
		data = append([]byte{}, item...)
		return nil
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (kv *KVObject) DBSaveState(key []byte, data []byte) error {
	return kv.KVDB.Update(func(txn *bolt.Tx) error {
		err := txn.Bucket([]byte("default")).Put(key, data)
		if err != nil {
			return fmt.Errorf("could not insert weight: %v", err)
		}
		return nil
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
