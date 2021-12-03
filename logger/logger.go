package logger

import "go.uber.org/zap"

// Logger ...
var Logger *zap.Logger

func init() {
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
}
