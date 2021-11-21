package barfeed

import (
	"goalgotrade/common"
	"time"

	"github.com/go-gota/gota/series"
)

type BarFilter interface {
	IncludeBar(bar common.Bar) bool
}

type CSVBarFeed struct {
	memBarFeed
	DailyBarTime   *time.Time
	BarFilter      BarFilter
	DateTimeFormat string
	ColumnNames    map[string]string
	HaveAdjClose   bool
}

func NewCSVBarFeed(freqs []common.Frequency, stype series.Type, timezone string, maxlen int) *CSVBarFeed {
	m := NewMemBarFeed(freqs, stype, maxlen)
	return &CSVBarFeed{
		memBarFeed:     *m,
		DateTimeFormat: "%Y-%m-%d %H:%M:%S",
		ColumnNames: map[string]string{
			"datetime":  "Date Time",
			"open":      "Open",
			"high":      "High",
			"low":       "Low",
			"close":     "Close",
			"volume":    "Volume",
			"adj_close": "Adj Close",
		},
		HaveAdjClose: false,
	}
}

func (c *CSVBarFeed) SetNoAdjClose() {
	c.ColumnNames["adj_close"] = ""
	c.HaveAdjClose = false
}

func (c *CSVBarFeed) AddBarsFromCSV(instrument string, path string, timezone string, skipMalformedBars bool) error {
	// TODO: implement me
	panic("not implemented")
}
