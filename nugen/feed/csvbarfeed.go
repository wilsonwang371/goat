package feed

import (
	"encoding/csv"
	"fmt"
	"goalgotrade/nugen/bar"
	"goalgotrade/nugen/consts/frequency"
	lg "goalgotrade/nugen/logger"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/go-gota/gota/series"
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

// CSVBarFeed ...
type CSVBarFeed interface {
	MemBarFeed
	IncludeBar(bar bar.Bar) bool
	AddBarsFromCSV(f CSVBarFeed, instrument string, path string, _ string) error
	SetNoAdjClose(f CSVBarFeed) error
}

type csvBarFeed struct {
	memBarFeed
	DailyBarTime   *time.Time
	DateTimeFormat string
	ColumnNames    map[ColumnName]string
	HaveAdjClose   bool
	TimeZone       string
}

// NewCSVBarFeed ...
func NewCSVBarFeed(freqList []frequency.Frequency, sType series.Type, timezone string, maxLen int) CSVBarFeed {
	return newCSVBarFeed(freqList, sType, timezone, maxLen)
}

func newCSVBarFeed(freqList []frequency.Frequency, sType series.Type, timezone string, maxLen int) *csvBarFeed {
	if len(freqList) != 1 {
		panic("currently csv bar feed only supports one frequency")
	}
	res := &csvBarFeed{
		memBarFeed:     *newMemBarFeed(freqList, sType, maxLen),
		DateTimeFormat: "%Y-%m-%d %H:%M:%S",
		ColumnNames: map[ColumnName]string{
			ColumnDateTime: "Date Time",
			ColumnOpen:     "Open",
			ColumnHigh:     "High",
			ColumnLow:      "Low",
			ColumnClose:    "Close",
			ColumnVolume:   "Volume",
			ColumnAdjClose: "Adj Close",
		},
		HaveAdjClose: false,
		TimeZone:     timezone,
	}
	return res
}

// IncludeBar ...
func (c *csvBarFeed) IncludeBar(bar bar.Bar) bool {
	panic("implement me")
}

// SetNoAdjClose ...
func (c *csvBarFeed) SetNoAdjClose(f CSVBarFeed) error {
	c.ColumnNames["adj_close"] = ""
	c.HaveAdjClose = false
	return c.SetUseAdjustedValue(f, false)
}

func (c *csvBarFeed) parseRawToBar(dict map[string]string) (bar.Bar, error) {
	dateTimeRaw := dict[c.ColumnNames[ColumnDateTime]]
	openRaw := dict[c.ColumnNames[ColumnOpen]]
	highRaw := dict[c.ColumnNames[ColumnHigh]]
	lowRaw := dict[c.ColumnNames[ColumnLow]]
	closeRaw := dict[c.ColumnNames[ColumnClose]]
	volumeRaw := dict[c.ColumnNames[ColumnVolume]]
	adjCloseRaw := ""
	if val, ok := dict[c.ColumnNames[ColumnAdjClose]]; ok {
		adjCloseRaw = val
	}
	if adjCloseRaw != "" {
		c.HaveAdjClose = true
	}
	carbonResult := carbon.ParseByFormat(c.DateTimeFormat, dateTimeRaw)
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
	bar, err := bar.NewBasicBar(dateTime, open, high, low, closeVal, volume, adjClose, c.frequencies[0])
	if err != nil {
		return nil, err
	}
	if c.HaveAdjClose {
		if err := bar.SetUseAdjustedValue(true); err != nil {
			return nil, err
		}
	}
	return bar, nil
}

// AddBarsFromCSV ...
func (c *csvBarFeed) AddBarsFromCSV(f CSVBarFeed, instrument string, path string, _ string) error {
	isHeader := true
	var headers []string

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := csv.NewReader(file)
	var loadedBarList []bar.Bar
	for {
		// Read each record from csv
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			lg.Logger.Error("read error", zap.Error(err))
			return err
		}

		if isHeader {
			headers = record
			isHeader = false
		} else {
			if headers == nil {
				return fmt.Errorf("invalid headers")
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
				return err
			}
			if bar != nil && f.IncludeBar(bar) {
				loadedBarList = append(loadedBarList, bar)
			}
		}
		if err := c.AddBarListFromSequence(instrument, loadedBarList); err != nil {
			return err
		}
	}
	return nil
}
