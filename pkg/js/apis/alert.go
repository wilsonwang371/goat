package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"
	"goat/pkg/notify"

	"github.com/dop251/goja"
	"go.uber.org/zap"
)

type AlertObject struct {
	cfg       *config.Config
	VM        *goja.Runtime
	notifiers []notify.Notifier
}

func NewAlertObject(cfg *config.Config, vm *goja.Runtime) (*AlertObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	alert := &AlertObject{
		cfg: cfg,
		VM:  vm,
	}

	alertObj := alert.VM.NewObject()
	alertObj.Set("info", alert.InfoCmd)
	alertObj.Set("warn", alert.WarnCmd)
	alertObj.Set("error", alert.ErrorCmd)
	alertObj.Set("email_msg", alert.EmailCmd)
	alertObj.Set("mobile_msg", alert.MobileCmd)
	if err := alert.VM.Set("alert", alertObj); err != nil {
		return nil, err
	}

	if cfg.Notification.Pushover.Enabled {
		alert.notifiers = append(alert.notifiers, notify.NewPushoverNotifier(cfg))
	}

	if cfg.Notification.Email.Enabled {
		alert.notifiers = append(alert.notifiers, notify.NewEmailNotifier(cfg))
	}

	if cfg.Notification.Twilio.Enabled {
		alert.notifiers = append(alert.notifiers, notify.NewTwilioNotifier(cfg))
	}

	return alert, nil
}

func (a *AlertObject) parseArgs(call goja.FunctionCall) (title, msg string, err error) {
	if len(call.Arguments) > 2 {
		return "", "", fmt.Errorf("invalid number of arguments")
	}
	if len(call.Arguments) == 1 {
		title = "<empty>"
		msg = call.Argument(0).String()
	} else {
		title = call.Argument(0).String()
		msg = call.Argument(1).String()
	}
	return title, msg, nil
}

func (a *AlertObject) InfoCmd(call goja.FunctionCall) goja.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		logger.Logger.Error("failed to parse args: %s", zap.Error(err))
		return a.VM.ToValue(false)
	} else {
		for _, n := range a.notifiers {
			if n.Level() <= config.InfoLevel {
				n.SetSubject(title)
				n.SetContent(msg)
				if err := n.Send(); err != nil {
					logger.Logger.Error("failed to send notification: %s", zap.Error(err))
				}
			}
		}
	}

	if errorCount > 0 {
		return a.VM.ToValue(false)
	}
	return a.VM.ToValue(true)
}

func (a *AlertObject) WarnCmd(call goja.FunctionCall) goja.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		logger.Logger.Error("failed to parse args: %s", zap.Error(err))
		return a.VM.ToValue(false)
	} else {
		for _, n := range a.notifiers {
			if n.Level() <= config.WarnLevel {
				n.SetSubject(title)
				n.SetContent(msg)
				if err := n.Send(); err != nil {
					logger.Logger.Error("failed to send notification: %s", zap.Error(err))
					errorCount++
				}
			}
		}
	}

	if errorCount > 0 {
		return a.VM.ToValue(false)
	}
	return a.VM.ToValue(true)
}

func (a *AlertObject) ErrorCmd(call goja.FunctionCall) goja.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		logger.Logger.Error("failed to parse args: %s", zap.Error(err))
		return a.VM.ToValue(false)
	} else {
		for _, n := range a.notifiers {
			if n.Level() <= config.ErrorLevel {
				n.SetSubject(title)
				n.SetContent(msg)
				if err := n.Send(); err != nil {
					logger.Logger.Error("failed to send notification: %s", zap.Error(err))
					errorCount++
				}
			}
		}
	}

	if errorCount > 0 {
		return a.VM.ToValue(false)
	}
	return a.VM.ToValue(true)
}

func (a *AlertObject) EmailCmd(call goja.FunctionCall) goja.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		logger.Logger.Error("failed to parse args: %s", zap.Error(err))
		return a.VM.ToValue(false)
	} else {
		for _, n := range a.notifiers {
			if n.FeatureFlags()&config.NotifyIsEmailFlag != 0 {
				n.SetSubject(title)
				n.SetContent(msg)
				if err := n.Send(); err != nil {
					logger.Logger.Error("failed to send notification: %s", zap.Error(err))
					errorCount++
				}
			}
		}
	}

	if errorCount > 0 {
		return a.VM.ToValue(false)
	}
	return a.VM.ToValue(true)
}

func (a *AlertObject) MobileCmd(call goja.FunctionCall) goja.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		logger.Logger.Error("failed to parse args: %s", zap.Error(err))
		return a.VM.ToValue(false)
	} else {
		for _, n := range a.notifiers {
			if n.FeatureFlags()&config.NotifyIsMobileFlag != 0 {
				n.SetSubject(title)
				n.SetContent(msg)
				if err := n.Send(); err != nil {
					logger.Logger.Error("failed to send notification: %s", zap.Error(err))
					errorCount++
				}
			}
		}
	}

	if errorCount > 0 {
		return a.VM.ToValue(false)
	}
	return a.VM.ToValue(true)
}
