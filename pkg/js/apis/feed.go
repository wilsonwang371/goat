package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/core"
	"goat/pkg/logger"

	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

type FeedObject struct {
	cfg  *config.Config
	feed core.DataFeed
	VM   *otto.Otto
}

func NewFeedObject(cfg *config.Config, vm *otto.Otto, f core.DataFeed) (*FeedObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	feed := &FeedObject{
		cfg:  cfg,
		feed: f,
		VM:   vm,
	}

	feedObj, err := feed.VM.Object(`feed = {}`)
	if err != nil {
		return nil, err
	}
	feedObj.Set("dataseries", feed.DataSeriesCmd)

	freqObj, err := feed.VM.Object(`frequency = {}`)
	if err != nil {
		return nil, err
	}
	freqObj.Set("REALTIME", core.REALTIME)
	freqObj.Set("SECOND", core.SECOND)
	freqObj.Set("MINUTE", core.MINUTE)
	freqObj.Set("HOUR", core.HOUR)
	freqObj.Set("HOUR_4", core.HOUR_4)
	freqObj.Set("DAY", core.DAY)
	freqObj.Set("WEEK", core.WEEK)
	freqObj.Set("MONTH", core.MONTH)
	freqObj.Set("YEAR", core.YEAR)

	return feed, nil
}

func (f *FeedObject) DataSeriesCmd(call otto.FunctionCall) otto.Value {
	var err error
	var symbol string
	var length int64
	var freq int64

	if len(call.ArgumentList) != 3 {
		logger.Logger.Debug("DataSeriesCmd needs 3 arguments")
		return otto.NullValue()
	}

	symbol = call.Argument(0).String()
	if symbol == "" {
		logger.Logger.Debug("invalid symbol")
		return otto.NullValue()
	}

	length, err = call.Argument(2).ToInteger()
	if err != nil || length <= 0 {
		logger.Logger.Error("invalid length")
		return otto.NullValue()
	}

	if freq, err = call.Argument(1).ToInteger(); err == nil {
		switch core.Frequency(freq) {
		case core.REALTIME, core.SECOND, core.MINUTE, core.HOUR, core.HOUR_4, core.DAY, core.WEEK, core.MONTH, core.YEAR:
			if f.feed == nil {
				logger.Logger.Error("feed is nil")
				return otto.NullValue()
			}
			if ds, err := f.feed.GetDataSeries(symbol, core.Frequency(freq)); err != nil {
				logger.Logger.Info("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
				return otto.NullValue()
			} else {
				if obj, err := ds.GetDataAsObjects(int(length)); err != nil {
					logger.Logger.Info("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
					return otto.NullValue()
				} else {
					if ret, err := f.VM.ToValue(obj); err != nil {
						logger.Logger.Info("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
						return otto.NullValue()
					} else {
						return ret
					}
				}
			}
		default:
			logger.Logger.Error("invalid frequency")
			return otto.NullValue()
		}
	} else {
		logger.Logger.Debug("invalid frequency")
		return otto.NullValue()
	}
}
