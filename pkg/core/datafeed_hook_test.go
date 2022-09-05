package core

import (
	"fmt"
	"math"
	"testing"
	"time"
)

const float64EqualityThreshold = 1e-2

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

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

	v = ctrl.PossibleOneNewValue()
	if v != nil {
		t.Error("should be nil")
	}
}

func TestHookVerify(t *testing.T) {
	ctrl := NewDataFeedValueHookControl()
	ctrl.AddNewHook(NewDayBarGenHook())

	for i := 1; i < 10; i++ {
		tm := time.Date(2016, time.January, 1, i, 0, 0, 0, time.UTC)
		barData := NewBasicBar(
			tm,
			100*float64(i),
			100*float64(i),
			100*float64(i),
			100*float64(i),
			100*float64(i),
			0, REALTIME)
		ctrl.FilterNewValue(&PendingDataFeedValue{
			t: tm,
			f: REALTIME,
			v: map[string]interface{}{
				"test": barData.(interface{}),
			},
		}, false)
		barData2 := NewBasicBar(
			tm,
			-100*float64(i),
			-100*float64(i),
			-100*float64(i),
			-100*float64(i),
			-100*float64(i),
			0, REALTIME)
		ctrl.FilterNewValue(&PendingDataFeedValue{
			t: tm,
			f: REALTIME,
			v: map[string]interface{}{
				"test": barData2.(interface{}),
			},
		}, false)
	}

	v := ctrl.PossibleOneNewValue()
	if v == nil {
		t.Error("should not be nil")
	}

	res := v.v["test"].(Bar)

	if !almostEqual(res.Open(), 100.000) || !almostEqual(res.Close(), -900.000) ||
		!almostEqual(res.High(), 900.000) || !almostEqual(res.Low(), -900.000) {
		tmp := fmt.Sprintf("open %f, close %f, high %f, low %f\n", res.Open(), res.Close(), res.High(), res.Low())
		t.Error("verify failed", tmp)
	}
}
