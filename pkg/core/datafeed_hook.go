package core

type DataFeedHooks interface {
	FilterNewValue(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue
	AddNewFilter(hook DataFeedHookFunc)
}

type DataFeedHookFunc func(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue

type dataFeedHook struct {
	hooks []DataFeedHookFunc
}

func NewDataFeedValueHook() DataFeedHooks {
	return &dataFeedHook{
		hooks: make([]DataFeedHookFunc, 0),
	}
}

func (d *dataFeedHook) FilterNewValue(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue {
	values := []*PendingDataFeedValue{value}
	for _, h := range d.hooks {
		for _, v := range h(value, isRecovery) {
			values = append(values, v)
		}
	}
	return values
}

func (d *dataFeedHook) AddNewFilter(hook DataFeedHookFunc) {
	d.hooks = append(d.hooks, hook)
}

func DayBarGenHook(value *PendingDataFeedValue, isRecovery bool) []*PendingDataFeedValue {
	return []*PendingDataFeedValue{value}
}
