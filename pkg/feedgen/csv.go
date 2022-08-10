package feedgen

import (
	"encoding/csv"
	"goalgotrade/pkg/core"
	lg "goalgotrade/pkg/logger"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/golang-module/carbon"
	"go.uber.org/zap"
)

// ColumnName ...
type ColumnName string

// ColumnDateTime ...
const (
	ColumnDateTime ColumnName = "dateTime"
	ColumnOpen     ColumnName = "open"
	ColumnHigh     ColumnName = "high"
	ColumnLow      ColumnName = "low"
	ColumnClose    ColumnName = "close"
	ColumnVolume   ColumnName = "volume"
	ColumnAdjClose ColumnName = "adj_close"
)

type CSVFeedGenerator struct {
	barfeed        core.FeedGenerator
	path           string
	dateTimeFormat string
	columnNames    map[ColumnName]string
	haveAdjClose   bool
	frequency      core.Frequency
	instrument     string
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
	panic("unimplemented")
}

func NewCSVBarFeedGenerator(path string, instrument string, freq core.Frequency) core.FeedGenerator {
	c := &CSVFeedGenerator{
		barfeed:        core.NewBarFeedGenerator([]core.Frequency{freq}, 100),
		path:           path,
		dateTimeFormat: "%Y-%m-%d %H:%M:%S",
		columnNames: map[ColumnName]string{
			ColumnDateTime: "Date Time",
			ColumnOpen:     "Open",
			ColumnHigh:     "High",
			ColumnLow:      "Low",
			ColumnClose:    "Close",
			ColumnVolume:   "Volume",
			ColumnAdjClose: "Adj Close",
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
		lg.Logger.Error("open file failed", zap.Error(err))
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
			lg.Logger.Error("read error", zap.Error(err))
			os.Exit(1)
		}

		if isHeader {
			headers = record
			isHeader = false
		} else {
			if headers == nil {
				lg.Logger.Error("invalid headers")
				os.Exit(1)
			}
			data := map[string]string{}
			for i, v := range record {
				if i < len(headers) {
					data[headers[i]] = v
				} else {
					lg.Logger.Warn("header not found", zap.Int("index", i), zap.String("value", v))
				}
			}
			bar, err := c.parseRawToBar(data)
			if err != nil {
				lg.Logger.Error("parse error", zap.Error(err))
				os.Exit(1)
			}
			c.AppendNewValueToBuffer(bar.DateTime(), map[string]interface{}{c.instrument: bar}, c.frequency)
		}
	}
	c.Finish()
}

func (c *CSVFeedGenerator) parseRawToBar(dict map[string]string) (core.Bar, error) {
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
	carbonResult := carbon.ParseByFormat(c.dateTimeFormat, dateTimeRaw)
	if carbonResult.Error != nil {
		return nil, carbonResult.Error
	}
	dateTime := carbonResult.Carbon2Time()
	open, err := strconv.ParseFloat(openRaw, 64)
	if err != nil {
		return nil, err
	}
	high, err := strconv.ParseFloat(highRaw, 64)
	if err != nil {
		return nil, err
	}
	low, err := strconv.ParseFloat(lowRaw, 64)
	if err != nil {
		return nil, err
	}
	closeVal, err := strconv.ParseFloat(closeRaw, 64)
	if err != nil {
		return nil, err
	}
	volume, err := strconv.ParseFloat(volumeRaw, 64)
	if err != nil {
		return nil, err
	}
	adjClose, err := strconv.ParseFloat(adjCloseRaw, 64)
	if err != nil {
		adjClose = .0
	}
	bar := core.NewBasicBar(dateTime, open, high, low, closeVal, adjClose, int64(volume), c.frequency)
	if c.haveAdjClose {
		bar.SetUseAdjustedValue(true)
	}
	return bar, nil
}
