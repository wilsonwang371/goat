package common

import "time"

const (
	LiveGenFailureSleepDuration = 10 * time.Second
	LiveGenFailureMaxCount      = 20

	ProfilePort = 6060
	MetricsPort = 2112

	IdleSleepDuration = 10 * time.Millisecond
)
