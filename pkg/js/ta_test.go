package js

import (
	"testing"

	"goalgotrade/pkg/logger"
)

func TestTALibMethods(t *testing.T) {
	methods := AllTALibMethods()
	for _, method := range methods {
		logger.Logger.Info(method.Name)
	}
}
