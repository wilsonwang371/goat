package core

type DataFeedHooksControl interface {
	FilterNewValue(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue
	AddNewHook(hook DataFeedHook)
}

type dataFeedHookControl struct {
	hooks []DataFeedHook
}

func NewDataFeedValueHookControl() DataFeedHooksControl {
	return &dataFeedHookControl{
		hooks: make([]DataFeedHook, 0),
	}
}

func (d *dataFeedHookControl) FilterNewValue(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue {
	values := []*PendingDataFeedValue{value}
	for _, h := range d.hooks {
		for _, v := range h.Invoke(value, isRecovery) {
			values = append(values, v)
		}
	}
	return values
}

func (d *dataFeedHookControl) AddNewHook(hook DataFeedHook) {
	d.hooks = append(d.hooks, hook)
}

type DataFeedHook interface {
	Invoke(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue
}

type dataFeedHook struct{}

func NewDayBarGenHook() DataFeedHook {
	return &dataFeedHook{}
}

// Invoke implements DataFeedHook
func (*dataFeedHook) Invoke(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue {
	return []*PendingDataFeedValue{value}
}
