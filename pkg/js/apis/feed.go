package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/core"
	"goat/pkg/logger"

	otto "github.com/dop251/goja"
	"go.uber.org/zap"
)

type FeedObject struct {
	cfg  *config.Config
	feed core.DataFeed
	VM   *otto.Runtime
}

func NewFeedObject(cfg *config.Config, vm *otto.Runtime, f core.DataFeed) (*FeedObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	feed := &FeedObject{
		cfg:  cfg,
		feed: f,
		VM:   vm,
	}

	feedObj := feed.VM.NewObject()
	feedObj.Set("dataseries", feed.DataSeriesCmd)
	if err := feed.VM.Set("feed", feedObj); err != nil {
		logger.Logger.Fatal("failed to set feed object", zap.Error(err))
		return nil, err
	}

	freqObj := feed.VM.NewObject()
	freqObj.Set("REALTIME", core.REALTIME)
	freqObj.Set("SECOND", core.SECOND)
	freqObj.Set("MINUTE", core.MINUTE)
	freqObj.Set("HOUR", core.HOUR)
	freqObj.Set("HOUR_4", core.HOUR_4)
	freqObj.Set("DAY", core.DAY)
	freqObj.Set("WEEK", core.WEEK)
	freqObj.Set("MONTH", core.MONTH)
	freqObj.Set("YEAR", core.YEAR)
	if err := feed.VM.Set("frequency", freqObj); err != nil {
		logger.Logger.Fatal("failed to set frequency object", zap.Error(err))
		return nil, err
	}

	return feed, nil
}

func (f *FeedObject) DataSeriesCmd(call otto.FunctionCall) otto.Value {
	var symbol string
	var length int64
	var freq int64

	if len(call.Arguments) != 3 {
		logger.Logger.Debug("DataSeriesCmd needs 3 arguments")
		return otto.Null()
	}

	symbol = call.Argument(0).String()
	if symbol == "" {
		logger.Logger.Debug("invalid symbol")
		return otto.Null()
	}

	length = call.Argument(2).ToInteger()

	freq = call.Argument(1).ToInteger()
	switch core.Frequency(freq) {
	case core.REALTIME, core.SECOND, core.MINUTE, core.HOUR, core.HOUR_4, core.DAY, core.WEEK, core.MONTH, core.YEAR:
		if f.feed == nil {
			logger.Logger.Error("feed is nil")
			return otto.Null()
		}
		if ds, err := f.feed.GetDataSeries(symbol, core.Frequency(freq)); err != nil {
			logger.Logger.Info("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
			return otto.Null()
		} else {
			if obj, err := ds.GetDataAsObjects(int(length)); err != nil {
				logger.Logger.Info("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
				return otto.Null()
			} else {
				return f.VM.ToValue(obj)
			}
		}
	default:
		logger.Logger.Error("invalid frequency")
		return otto.Null()
	}
}
