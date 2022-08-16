package feedgen

import (
	"fmt"
	"time"

	"goalgotrade/pkg/core"
	"goalgotrade/pkg/logger"

	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"go.uber.org/zap"
)

type YahooFeedGenerator struct {
	barfeed      core.FeedGenerator
	haveAdjClose bool
	frequency    core.Frequency
	instrument   string
}

// IsComplete implements core.FeedGenerator
func (y *YahooFeedGenerator) IsComplete() bool {
	return y.barfeed.IsComplete()
}

var freqMapping = map[core.Frequency]datetime.Interval{
	core.DAY:     datetime.OneDay,
	core.WEEK:    datetime.FiveDay,
	core.MONTH:   datetime.OneMonth,
	core.YEAR:    datetime.OneYear,
	core.UNKNOWN: datetime.OneDay, // by default we use one day
}

// AppendNewValueToBuffer implements core.FeedGenerator
func (*YahooFeedGenerator) AppendNewValueToBuffer(time.Time, map[string]interface{}, core.Frequency) error {
	panic("unimplemented")
}

// CreateDataSeries implements core.FeedGenerator
func (y *YahooFeedGenerator) CreateDataSeries(key string, maxLen int) core.DataSeries {
	return y.barfeed.CreateDataSeries(key, maxLen)
}

// Finish implements core.FeedGenerator
func (y *YahooFeedGenerator) Finish() {
	y.barfeed.Finish()
}

// PeekNextTime implements core.FeedGenerator
func (y *YahooFeedGenerator) PeekNextTime() *time.Time {
	return y.barfeed.PeekNextTime()
}

// PopNextValues implements core.FeedGenerator
func (y *YahooFeedGenerator) PopNextValues() (time.Time, map[string]interface{}, core.Frequency, error) {
	return y.barfeed.PopNextValues()
}

func NewYahooBarFeedGenerator(instrument string, freq core.Frequency) core.FeedGenerator {
	rtn := &YahooFeedGenerator{
		barfeed:      core.NewBarFeedGenerator([]core.Frequency{freq}, 100),
		haveAdjClose: true,
		frequency:    freq,
		instrument:   instrument,
	}

	params := &chart.Params{
		Symbol:   instrument,
		Start:    datetime.FromUnix(int(time.Now().AddDate(-1, 0, 0).Unix())),
		End:      datetime.FromUnix(int(time.Now().Unix())),
		Interval: freqMapping[freq],
	}
	if interval, ok := freqMapping[freq]; ok {
		params.Interval = interval
	} else {
		logger.Logger.Info("unknown frequency, use default one day", zap.String("instrument", instrument),
			zap.String("frequency", fmt.Sprintf("%v", freq)))
		params.Interval = datetime.OneDay
	}

	go func() {
		iter := chart.Get(params)

		for iter.Next() {
			// logger.Logger.Info("yahoo", zap.String("bar", fmt.Sprintf("%+v", iter.Bar())))
			ts := iter.Bar().Timestamp
			err := rtn.barfeed.AppendNewValueToBuffer(time.Unix(int64(ts), 0), map[string]interface{}{
				rtn.instrument: core.NewBasicBar(
					time.Unix(int64(ts), 0),
					iter.Bar().Open.InexactFloat64(),
					iter.Bar().High.InexactFloat64(),
					iter.Bar().Low.InexactFloat64(),
					iter.Bar().Close.InexactFloat64(),
					iter.Bar().AdjClose.InexactFloat64(),
					int64(iter.Bar().Volume), rtn.frequency),
			}, rtn.frequency)
			if err != nil {
				logger.Logger.Error("yahoo", zap.Error(err))
				panic(err)
			}
		}

		rtn.Finish()

		if err := iter.Err(); err != nil {
			logger.Logger.Error("yahoo", zap.Error(err))
			return
		}
	}()

	return rtn
}
