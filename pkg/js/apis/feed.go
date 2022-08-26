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
		logger.Logger.Debug("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq))
		switch core.Frequency(freq) {
		case core.REALTIME, core.SECOND, core.MINUTE, core.HOUR, core.HOUR_4, core.DAY, core.WEEK, core.MONTH, core.YEAR:
			logger.Logger.Debug("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq))
			if f.feed == nil {
				logger.Logger.Error("feed is nil")
				return otto.NullValue()
			}
			if ds, err := f.feed.GetDataSeries(symbol, core.Frequency(freq)); err != nil {
				logger.Logger.Error("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
				return otto.NullValue()
			} else {
				if obj, err := ds.GetObject(); err != nil {
					logger.Logger.Error("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
					return otto.NullValue()
				} else {
					if ret, err := f.VM.ToValue(obj); err != nil {
						logger.Logger.Error("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Error(err))
						return otto.NullValue()
					} else {
						logger.Logger.Debug("DataSeriesCmd", zap.String("symbol", symbol), zap.Int64("freq", freq), zap.Any("ret", ret))
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
