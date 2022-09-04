package core

import "time"

type DataFeedHooksControl interface {
	FilterNewValue(value *PendingDataFeedValue, isRecovery bool)
	PossibleOneNewValue() *PendingDataFeedValue
	AddNewHook(hook DataFeedHook)
}

type dataFeedHookControl struct {
	hooks []DataFeedHook
}

// PossibleOneNewValue implements DataFeedHooksControl
func (d *dataFeedHookControl) PossibleOneNewValue() *PendingDataFeedValue {
	for _, h := range d.hooks {
		if v := h.MayHaveNewValue(); v != nil {
			return v
		}
	}
	return nil
}

func NewDataFeedValueHookControl() DataFeedHooksControl {
	return &dataFeedHookControl{
		hooks: make([]DataFeedHook, 0),
	}
}

func (d *dataFeedHookControl) FilterNewValue(value *PendingDataFeedValue, isRecovery bool) {
	for _, h := range d.hooks {
		h.Invoke(value, isRecovery)
	}
}

func (d *dataFeedHookControl) AddNewHook(hook DataFeedHook) {
	d.hooks = append(d.hooks, hook)
}

type DataFeedHook interface {
	Invoke(value *PendingDataFeedValue, isRecovery bool)
	MayHaveNewValue() *PendingDataFeedValue
}

type dataFeedHook struct {
	dayBarMap map[string]Bar
	startTime *time.Time
	stopTime  *time.Time
}

func NewDayBarGenHook() DataFeedHook {
	return &dataFeedHook{
		dayBarMap: make(map[string]Bar),
		startTime: nil,
		stopTime:  nil,
	}
}

func (d *dataFeedHook) timeToGenDayBar() bool {
	if d.startTime == nil {
		return false
	}
	if time.Now().UTC().After(*d.stopTime) {
		return true
	}
	return false
}

func (d *dataFeedHook) MayHaveNewValue() *PendingDataFeedValue {
	if d.timeToGenDayBar() {
		var newDayBar PendingDataFeedValue
		newDayBar.f = DAY
		newDayBar.v = map[string]interface{}{}
		for k, v := range d.dayBarMap {
			newDayBar.v[k] = v
		}
		newDayBar.t = time.Date(d.startTime.Year(), d.startTime.Month(), d.startTime.Day(), 0, 0, 0, 0, time.UTC)
		d.startTime = nil
		d.stopTime = nil
		d.dayBarMap = make(map[string]Bar)
		return &newDayBar
	}
	return nil
}

// Invoke implements DataFeedHook
func (d *dataFeedHook) Invoke(value *PendingDataFeedValue, isRecovery bool) {
	if isRecovery {
		return
	}
	if !(value.f >= REALTIME || value.f < DAY) {
		// we don't care about other data frequencies
		return
	}
	if d.startTime == nil {
		d.startTime = &value.t
		// set the stop time
		stopTime := time.Date(d.startTime.Year(), d.startTime.Month(), d.startTime.Day(), 23, 59, 59, 0, time.UTC)
		d.stopTime = &stopTime
	}
	for k, v := range value.v {
		if d.dayBarMap[k] == nil {
			bar := v.(Bar)
			d.dayBarMap[k] = NewBasicBar(value.t, bar.Open(), bar.High(), bar.Low(), bar.Close(),
				bar.AdjClose(), bar.Volume(), DAY)
		} else {
			bar := v.(Bar)
			if bar.Close() > d.dayBarMap[k].High() {
				d.dayBarMap[k] = NewBasicBar(d.dayBarMap[k].DateTime(), d.dayBarMap[k].Open(), bar.Close(), d.dayBarMap[k].Low(), bar.Close(),
					d.dayBarMap[k].AdjClose(), d.dayBarMap[k].Volume(), d.dayBarMap[k].Frequency())
			}
			if bar.Close() < d.dayBarMap[k].Low() {
				d.dayBarMap[k] = NewBasicBar(d.dayBarMap[k].DateTime(), d.dayBarMap[k].Open(), d.dayBarMap[k].High(), bar.Close(), bar.Close(),
					d.dayBarMap[k].AdjClose(), d.dayBarMap[k].Volume(), d.dayBarMap[k].Frequency())
			}
		}
	}
	// if d.timeToGenDayBar() {
	// 	var newDayBar PendingDataFeedValue
	// 	newDayBar.f = DAY
	// 	newDayBar.v = map[string]interface{}{}
	// 	for k, v := range d.dayBarMap {
	// 		newDayBar.v[k] = v
	// 	}
	// 	newDayBar.t = time.Date(d.startTime.Year(), d.startTime.Month(), d.startTime.Day(), 0, 0, 0, 0, time.UTC)
	// 	d.startTime = nil
	// 	d.dayBarMap = make(map[string]Bar)
	// 	return []*PendingDataFeedValue{value, &newDayBar}
	// }
	// return []*PendingDataFeedValue{value}
}
