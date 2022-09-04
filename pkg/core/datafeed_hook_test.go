package core

import (
	"fmt"
	"testing"
	"time"
)

func TestHookSimple(t *testing.T) {
	ctrl := NewDataFeedValueHookControl()
	ctrl.AddNewHook(NewDayBarGenHook())

	barData := NewBasicBar(time.Now().UTC(),
		1.0, 1.0, 1.0, 1.0, 1.0, 0, REALTIME)

	ctrl.FilterNewValue(&PendingDataFeedValue{
		t: time.Now().UTC(),
		f: REALTIME,
		v: map[string]interface{}{
			"test": barData.(interface{}),
		},
	}, false)

	v := ctrl.PossibleOneNewValue()
	fmt.Printf("%+v\n", v)
	if v != nil {
		t.Error("should be nil")
	}
}

func TestHookSimple2(t *testing.T) {
	ctrl := NewDataFeedValueHookControl()
	ctrl.AddNewHook(NewDayBarGenHook())
	tm := time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)

	barData := NewBasicBar(tm,
		1.0, 1.0, 1.0, 1.0, 1.0, 0, REALTIME)

	ctrl.FilterNewValue(&PendingDataFeedValue{
		t: tm,
		f: REALTIME,
		v: map[string]interface{}{
			"test": barData.(interface{}),
		},
	}, false)

	v := ctrl.PossibleOneNewValue()
	fmt.Printf("%+v\n", v)
	if v == nil {
		t.Error("should not be nil")
	}
}
