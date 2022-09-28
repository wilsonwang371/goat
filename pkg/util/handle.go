package util

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"goat/pkg/logger"
	"goat/pkg/notify"

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

func PanicHandler(n notify.Notifier) {
	if r := recover(); r != nil {
		logger.Logger.Error("panic", zap.Any("panic", r))
		n.SetSubject("PANIC")
		n.SetContent(fmt.Sprintf("PANIC: %v", r))
		if err := n.Send(); err != nil {
			logger.Logger.Error("failed to send panic notification", zap.Error(err))
			return
		}
	}
}
