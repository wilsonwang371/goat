package apis

import (
	"fmt"

	"goat/pkg/config"
	"goat/pkg/logger"

	"github.com/robertkrimen/otto"
)

type FeedObject struct {
	cfg *config.Config
	VM  *otto.Otto
}

func NewFeedObject(cfg *config.Config, vm *otto.Otto) (*FeedObject, error) {
	if cfg == nil || vm == nil {
		return nil, fmt.Errorf("invalid config or vm")
	}

	feed := &FeedObject{
		cfg: cfg,
		VM:  vm,
	}

	feedObj, err := feed.VM.Object(`feed = {}`)
	if err != nil {
		return nil, err
	}
	feedObj.Set("dataseries", feed.DataSeriesCmd)

	return feed, nil
}

func (f *FeedObject) DataSeriesCmd(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) != 2 {
		logger.Logger.Debug("DataSeriesCmd needs 2 arguments")
		return otto.NullValue()
	}

	// TODO: validate arguments and return data series values

	return otto.NullValue()
}
