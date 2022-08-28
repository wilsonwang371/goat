package consts

import "time"

const (
	LiveGenFailureSleepDuration = 10 * time.Second
	LiveGenFailureMaxCount      = 20

	ProfilePort = 6060

	IdelSleepDuration = 10 * time.Millisecond
)
