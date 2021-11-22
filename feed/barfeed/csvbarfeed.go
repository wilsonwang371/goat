package barfeed

import (
	"encoding/csv"
	"fmt"
	"goalgotrade/common"
	"goalgotrade/core"
	lg "goalgotrade/logger"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/go-gota/gota/series"
	"github.com/golang-module/carbon"
	"go.uber.org/zap"
)

type ColumnName string

const (
	ColumnDateTime ColumnName = "dateTime"
	ColumnOpen     ColumnName = "open"
	ColumnHigh     ColumnName = "high"
	ColumnLow      ColumnName = "low"
	ColumnClose    ColumnName = "close"
	ColumnVolume   ColumnName = "volume"
	ColumnAdjClose ColumnName = "adj_close"
)

type BarFilter interface {
	IncludeBar(bar common.Bar) bool
}

type CSVBarFeed struct {
	memBarFeed
	DailyBarTime   *time.Time
	BarFilter      BarFilter
	DateTimeFormat string
	ColumnNames    map[ColumnName]string
	HaveAdjClose   bool
	TimeZone       string
}

func NewCSVBarFeed(freqList []common.Frequency, sType series.Type, timezone string, maxLen int) *CSVBarFeed {
	if len(freqList) != 1 {
		panic("currently csv barfeed only supports one frequency")
	}
	m := NewMemBarFeed(freqList, sType, maxLen)
	return &CSVBarFeed{
		memBarFeed:     *m,
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
}

func (c *CSVBarFeed) SetNoAdjClose() {
	c.ColumnNames["adj_close"] = ""
	c.HaveAdjClose = false
}

func (c *CSVBarFeed) parseRawToBar(dict map[string]string) (common.Bar, error) {
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
	bar := core.NewBasicBar(dateTime, open, high, low, closeVal, volume, adjClose, c.frequencies[0])
	if c.HaveAdjClose {
		if err := bar.SetUseAdjustedValue(true); err != nil {
			return nil, err
		}
	}
	return bar, nil
}

func (c *CSVBarFeed) AddBarsFromCSV(instrument string, path string, _ string) error {
	isHeader := true
	var headers []string

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := csv.NewReader(file)
	var loadedBarList []common.Bar
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
			if bar != nil && (c.BarFilter == nil || c.BarFilter.IncludeBar(bar)) {
				loadedBarList = append(loadedBarList, bar)
			}
		}
		if err := c.AddBarListFromSequence(instrument, loadedBarList); err != nil {
			return err
		}
	}
	return nil
}
