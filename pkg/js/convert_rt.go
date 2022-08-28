package js

import (
	"strconv"
	"time"

	"goat/pkg/config"
	"goat/pkg/convert"
	"goat/pkg/db"
	"goat/pkg/js/apis"
	"goat/pkg/logger"

	"github.com/schollz/progressbar/v3"

	"github.com/araddon/dateparse"
	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

const dbBatchCreateSize = 2048

type ConvertRuntime interface {
	Compile(source string) (*otto.Script, error)
	Execute(script *otto.Script) (otto.Value, error)
	Convert(dbsource convert.DBSource, dboutput *db.DB) error
}

type convertRt struct {
	cfg     *config.Config
	vm      *otto.Otto
	mapping *apis.DBMappingObject
	bar     *progressbar.ProgressBar
}

// Compile implements ConvertRuntime
func (c *convertRt) Compile(source string) (*otto.Script, error) {
	compiled, err := c.vm.Compile("", source)
	if err != nil {
		return nil, err
	}
	return compiled, nil
}

// Convert implements ConvertRuntime
func (c *convertRt) Convert(dbsource convert.DBSource, dboutput *db.DB) error {
	if err := dbsource.Open(); err != nil {
		return err
	}
	defer dbsource.Close()

	var count int64 = 0
	mappings := c.mapping.Mappings
	logger.Logger.Debug("mappings", zap.Any("mappings", mappings))

	c.bar = progressbar.Default(dbsource.TotalCount())

	allbars := []*db.BarData{}
	for {
		row, err := dbsource.ReadOneRow()
		if err != nil {
			return err
		}
		if row == nil {
			break
		}
		// process data
		var datetime time.Time
		var open, high, low, close, volume, adj_close float64
		var frequency int64
		var symbol, note string

		if val, ok := row[mappings["symbol"].(string)]; ok {
			symbol = val
		}
		if val, ok := row[mappings["datetime"].(string)]; ok {
			if val, err := dateparse.ParseAny(val); err == nil {
				datetime = val
			} else {
				return err
			}
		}
		if val, ok := row[mappings["open"].(string)]; ok {
			if tmp, err := strconv.ParseFloat(val, 64); err == nil {
				open = tmp
			} else {
				return err
			}
		}
		if val, ok := row[mappings["high"].(string)]; ok {
			if tmp, err := strconv.ParseFloat(val, 64); err == nil {
				high = tmp
			} else {
				return err
			}
		}
		if val, ok := row[mappings["low"].(string)]; ok {
			if tmp, err := strconv.ParseFloat(val, 64); err == nil {
				low = tmp
			} else {
				return err
			}
		}
		if val, ok := row[mappings["close"].(string)]; ok {
			if tmp, err := strconv.ParseFloat(val, 64); err == nil {
				close = tmp
			} else {
				return err
			}
		}
		if val, ok := row[mappings["volume"].(string)]; ok {
			if tmp, err := strconv.ParseFloat(val, 64); err == nil {
				volume = tmp
			} else {
				return err
			}
		}
		if val, ok := row[mappings["adj_close"].(string)]; ok {
			if tmp, err := strconv.ParseFloat(val, 64); err == nil {
				adj_close = tmp
			} else {
				return err
			}
		}
		if val, ok := row[mappings["frequency"].(string)]; ok {
			if tmp, err := strconv.ParseInt(val, 10, 64); err == nil {
				frequency = tmp
			} else {
				return err
			}
		}
		if val, ok := row[mappings["note"].(string)]; ok {
			note = val
		}

		bar := &db.BarData{
			Symbol:    symbol,
			DateTime:  datetime.Unix(),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    int64(volume),
			AdjClose:  adj_close,
			Frequency: frequency,
			Note:      note,
		}
		// logger.Logger.Debug("bar", zap.Any("bar", bar))
		allbars = append(allbars, bar)
		if len(allbars) >= dbBatchCreateSize {
			res := dboutput.Create(allbars)
			if res.Error != nil {
				return res.Error
			}
			allbars = []*db.BarData{}
		}

		count++
		c.bar.Add(1)
	}
	if len(allbars) > 0 {
		res := dboutput.Create(allbars)
		if res.Error != nil {
			return res.Error
		}
	}
	c.bar.Finish()
	return nil
}

// Execute implements ConvertRuntime
func (c *convertRt) Execute(script *otto.Script) (otto.Value, error) {
	return c.vm.Run(script)
}

func NewDBConvertRuntime(cfg *config.Config) ConvertRuntime {
	var err error
	res := &convertRt{
		cfg: cfg,
		vm:  otto.New(),
		bar: nil,
	}

	res.mapping, err = apis.NewDBMappingObject(cfg, res.vm)
	if err != nil {
		logger.Logger.Error("failed to create db convert mapping object", zap.Error(err))
		panic(err)
	}

	return res
}
