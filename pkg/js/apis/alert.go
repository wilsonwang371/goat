package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"
	"goat/pkg/notify"

	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

type AlertObject struct {
	cfg       *config.Config
	VM        *otto.Otto
	notifiers []notify.Notifier
}

func NewAlertObject(cfg *config.Config, vm *otto.Otto) (*AlertObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	alert := &AlertObject{
		cfg: cfg,
		VM:  vm,
	}

	alertObj, err := alert.VM.Object(`alert = {}`)
	if err != nil {
		return nil, err
	}
	alertObj.Set("info", alert.InfoCmd)
	alertObj.Set("warn", alert.WarnCmd)
	alertObj.Set("error", alert.ErrorCmd)
	alertObj.Set("email_msg", alert.EmailCmd)
	alertObj.Set("mobile_msg", alert.MobileCmd)

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

func (a *AlertObject) parseArgs(call otto.FunctionCall) (title, msg string, err error) {
	if len(call.ArgumentList) != 2 {
		return "", "", fmt.Errorf("invalid number of arguments")
	}
	title = call.Argument(0).String()
	msg = call.Argument(1).String()
	return title, msg, nil
}

func (a *AlertObject) InfoCmd(call otto.FunctionCall) otto.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		return otto.FalseValue()
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
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

func (a *AlertObject) WarnCmd(call otto.FunctionCall) otto.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		return otto.FalseValue()
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
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

func (a *AlertObject) ErrorCmd(call otto.FunctionCall) otto.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		return otto.FalseValue()
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
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

func (a *AlertObject) EmailCmd(call otto.FunctionCall) otto.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		return otto.FalseValue()
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
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

func (a *AlertObject) MobileCmd(call otto.FunctionCall) otto.Value {
	errorCount := 0

	if title, msg, err := a.parseArgs(call); err != nil {
		return otto.FalseValue()
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
		return otto.FalseValue()
	}
	return otto.TrueValue()
}
