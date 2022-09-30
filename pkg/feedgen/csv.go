package feedgen

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"time"

	"goat/pkg/core"
	"goat/pkg/logger"

	"github.com/araddon/dateparse"
	"github.com/golang-module/carbon"
	"go.uber.org/zap"
)

// ColumnName ...
type ColumnName string

// ColumnDateTime ...
const (
	ColumnDateTime  ColumnName = "dateTime"
	ColumnOpen      ColumnName = "open"
	ColumnHigh      ColumnName = "high"
	ColumnLow       ColumnName = "low"
	ColumnClose     ColumnName = "close"
	ColumnVolume    ColumnName = "volume"
	ColumnAdjClose  ColumnName = "adj_close"
	ColumnSymbol    ColumnName = "symbol"
	ColumnFrequency ColumnName = "frequency"
)

type CSVFeedGenerator struct {
	barfeed         core.FeedGenerator
	path            string
	dateTimeFormats []string
	columnNames     map[ColumnName]string
	haveAdjClose    bool
	frequency       core.Frequency
	instrument      string // default instrument name in case of no symbol column
}

// IsComplete implements core.FeedGenerator
func (c *CSVFeedGenerator) IsComplete() bool {
	return c.barfeed.IsComplete()
}

// AppendNewValueToBuffer implements core.FeedGenerator
func (c *CSVFeedGenerator) AppendNewValueToBuffer(time.Time, map[string]interface{}, core.Frequency) error {
	panic("unimplemented")
}

// CreateDataSeries implements core.FeedGenerator
func (c *CSVFeedGenerator) CreateDataSeries(key string, maxLen int) core.DataSeries {
	return c.barfeed.CreateDataSeries(key, maxLen)
}

// Finish implements core.FeedGenerator
func (c *CSVFeedGenerator) Finish() {
	c.barfeed.Finish()
}

// PeekNextTime implements core.FeedGenerator
func (c *CSVFeedGenerator) PeekNextTime() *time.Time {
	return c.barfeed.PeekNextTime()
}

// PopNextValues implements core.FeedGenerator
func (c *CSVFeedGenerator) PopNextValues() (time.Time, map[string]interface{}, core.Frequency, error) {
	return c.barfeed.PopNextValues()
}

func NewCSVBarFeedGenerator(path string, instrument string, freq core.Frequency) core.FeedGenerator {
	c := &CSVFeedGenerator{
		barfeed:         core.NewBarFeedGenerator([]core.Frequency{freq}, 100),
		path:            path,
		dateTimeFormats: []string{"%Y-%m-%d %H:%M:%S", "%Y-%m-%d"},
		columnNames: map[ColumnName]string{
			ColumnDateTime:  "Date",
			ColumnOpen:      "Open",
			ColumnHigh:      "High",
			ColumnLow:       "Low",
			ColumnClose:     "Close",
			ColumnVolume:    "Volume",
			ColumnAdjClose:  "Adj Close",
			ColumnSymbol:    "Symbol",
			ColumnFrequency: "Frequency",
		},
		haveAdjClose: false,
		frequency:    freq,
		instrument:   instrument,
	}
	go c.addBarsFromCSV()
	return c
}

func (c *CSVFeedGenerator) addBarsFromCSV() {
	isHeader := true
	var headers []string

	file, err := os.Open(c.path)
	if err != nil {
		logger.Logger.Error("open file failed", zap.Error(err))
		os.Exit(1)
	}

	reader := csv.NewReader(file)
	for {
		// Read each record from csv
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Logger.Error("read error", zap.Error(err))
			os.Exit(1)
		}

		if isHeader {
			headers = record
			isHeader = false
		} else {
			if headers == nil {
				logger.Logger.Error("invalid headers")
				os.Exit(1)
			}
			data := map[string]string{}
			for i, v := range record {
				if i < len(headers) {
					data[headers[i]] = v
				} else {
					logger.Logger.Warn("header not found", zap.Int("index", i), zap.String("value", v))
				}
			}
			symbol, bar, err := c.parseRawToBar(data)
			if err != nil {
				logger.Logger.Error("parse error", zap.Error(err))
				os.Exit(1)
			}
			for {
				if err := c.barfeed.AppendNewValueToBuffer(bar.DateTime(), map[string]interface{}{symbol: bar}, bar.Frequency()); err != nil {
					logger.Logger.Warn("append error", zap.Error(err))
					time.Sleep(time.Second)
				} else {
					break
				}
			}

		}
	}
	c.Finish()
}

func (c *CSVFeedGenerator) parseRawToBar(dict map[string]string) (string, core.Bar, error) {
	// logger.Logger.Info("parseRawToBar", zap.Any("dict", dict), zap.Any("columnNames", c.columnNames))
	dateTimeRaw := dict[c.columnNames[ColumnDateTime]]
	openRaw := dict[c.columnNames[ColumnOpen]]
	highRaw := dict[c.columnNames[ColumnHigh]]
	lowRaw := dict[c.columnNames[ColumnLow]]
	closeRaw := dict[c.columnNames[ColumnClose]]
	volumeRaw := dict[c.columnNames[ColumnVolume]]
	adjCloseRaw := ""
	if val, ok := dict[c.columnNames[ColumnAdjClose]]; ok {
		adjCloseRaw = val
	}
	if adjCloseRaw != "" {
		c.haveAdjClose = true
	}

	var symbol string
	if val, ok := dict[c.columnNames[ColumnSymbol]]; ok {
		symbol = val
	} else {
		symbol = c.instrument
	}

	var frequency core.Frequency
	if valStr, ok := dict[c.columnNames[ColumnFrequency]]; ok {
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			logger.Logger.Error("parse frequency error", zap.Error(err))
			os.Exit(1)
		} else {
			frequency = core.Frequency(val)
		}
	} else {
		frequency = c.frequency
	}

	dateTime := time.Time{}
	parseFailed := true
	for _, format := range c.dateTimeFormats {
		// logger.Logger.Info("parse date time", zap.String("raw", dateTimeRaw), zap.String("format", format))
		carbonResult := carbon.ParseByFormat(format, dateTimeRaw)
		if carbonResult.Error == nil {
			dateTime = carbonResult.Carbon2Time()
			parseFailed = false
			break
		}
	}

	if parseFailed {
		// logger.Logger.Debug("carbon failed, try dateparse", zap.String("dateTimeRaw", dateTimeRaw), zap.Error(carbonResult.Error))
		if val, err := dateparse.ParseAny(dateTimeRaw); err == nil {
			dateTime = val
		} else {
			// logger.Logger.Error("dateparse failed", zap.String("dateTimeRaw", dateTimeRaw), zap.Error(err))
			return "", nil, err
		}
	}
	open, err := strconv.ParseFloat(openRaw, 64)
	if err != nil {
		// logger.Logger.Error("parse open error", zap.Error(err))
		return "", nil, err
	}
	high, err := strconv.ParseFloat(highRaw, 64)
	if err != nil {
		// logger.Logger.Error("parse high error", zap.Error(err))
		return "", nil, err
	}
	low, err := strconv.ParseFloat(lowRaw, 64)
	if err != nil {
		// logger.Logger.Error("parse low error", zap.Error(err))
		return "", nil, err
	}
	closeVal, err := strconv.ParseFloat(closeRaw, 64)
	if err != nil {
		// logger.Logger.Error("parse close error", zap.Error(err))
		return "", nil, err
	}
	volume, err := strconv.ParseFloat(volumeRaw, 64)
	if err != nil {
		// logger.Logger.Error("parse volume error", zap.Error(err))
		return "", nil, err
	}
	adjClose, err := strconv.ParseFloat(adjCloseRaw, 64)
	if err != nil {
		// logger.Logger.Error("parse adjClose error", zap.Error(err))
		adjClose = .0
	}
	bar := core.NewBasicBar(dateTime, open, high, low, closeVal, adjClose, int64(volume), frequency)
	if c.haveAdjClose {
		bar.SetUseAdjustedValue(true)
	}
	return symbol, bar, nil
}
