package apis

import (
	"os"
	"testing"

	"goat/pkg/config"
	"goat/pkg/util"

	"github.com/dop251/goja"
)

func TestSimpleKV(t *testing.T) {
	ctx := util.NewTerminationContext()
	cfg := &config.Config{}
	os.RemoveAll("test.kvdb")
	defer os.RemoveAll("test.kvdb")
	obj, err := NewKVObject(ctx, cfg, goja.New(), "test.kvdb")
	if err != nil {
		t.Fatal(err)
	}
	if obj.DBSaveState([]byte("test"), []byte("test")) != nil {
		t.Fatal("failed to save state")
	}
	if res, err := obj.DBLoadState([]byte("test")); err != nil {
		t.Fatal("failed to get state", err)
	} else {
		if string(res) != "test" {
			t.Fatal("failed to get state", string(res))
		}
	}
}
