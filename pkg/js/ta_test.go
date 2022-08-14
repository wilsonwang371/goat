package js

import (
	"goalgotrade/pkg/logger"
	"testing"
)

func TestTALibMethods(t *testing.T) {
	methods := AllTALibMethods()
	for _, method := range methods {
		logger.Logger.Info(method.Name)
	}
}
