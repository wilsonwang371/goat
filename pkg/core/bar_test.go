package core

import (
	"testing"
	"time"
)

func TestSimpleBasicBar(t *testing.T) {
	b := NewBasicBar(time.Now(), .0, 2.0, 3.0, 1.2, 1.2, 100, REALTIME)
	b.SetMeta(BarMetaIsRecovery, true)
	if v := b.GetMeta(BarMetaIsRecovery); v != nil && v.(bool) {
		t.Log("ok")
	} else {
		t.Error("failed")
	}
}
