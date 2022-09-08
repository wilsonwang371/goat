package util

import (
	"context"
	"os"
	"os/signal"

	"goat/pkg/logger"

	"go.uber.org/zap"
)

func NewTerminationContext() context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		interrupt := <-c
		logger.Logger.Info("received interrupt", zap.String("interrupt", interrupt.String()))
		cancel()
	}()

	return ctx
}
