package db

import (
	"fmt"
	"os"

	"goat/pkg/logger"

	"go.uber.org/zap"
)

func MergeDBs(output *DB, sources []*DB) error {
	if output == nil {
		return fmt.Errorf("output db is nil")
	}
	if len(sources) == 0 {
		return fmt.Errorf("no input db")
	}

	var totalCount int64
	for _, source := range sources {
		count := source.FetchAll(true)
		totalCount += count
	}
	logger.Logger.Info("total count", zap.Int64("count", totalCount))

loopNext:
	for {
		var nextData *BarData
		nextIdx := -1
		for idx, oneSource := range sources {
			cmpNextData, err := oneSource.Peek()
			if err != nil {
				logger.Logger.Error("failed to peek data", zap.Error(err))
				os.Exit(1)
			}
			if cmpNextData == nil {
				if len(sources) == 1 {
					return nil
				}
				sources = append(sources[:idx], sources[idx+1:]...)
				continue loopNext
			}
			if nextData == nil || cmpNextData.DateTime < nextData.DateTime {
				nextData = cmpNextData
				nextIdx = idx
			}
		}
		bar := &BarData{
			Symbol:    nextData.Symbol,
			DateTime:  nextData.DateTime,
			Open:      nextData.Open,
			High:      nextData.High,
			Low:       nextData.Low,
			Close:     nextData.Close,
			Volume:    nextData.Volume,
			AdjClose:  nextData.AdjClose,
			Frequency: nextData.Frequency,
			Note:      nextData.Note,
		}
		output.Create(bar)
		sources[nextIdx].Next()
	}
	// we should never reach here
}
