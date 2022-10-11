package db

import (
	"fmt"
	"os"
	"time"

	"goat/pkg/logger"

	progressBar "github.com/schollz/progressbar/v3"

	"go.uber.org/zap"
)

var dbBatchCreateSize = 1024

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

	mergeBar := progressBar.Default(totalCount)
	pendingBars := []*BarData{}
	lastTime := time.Now().Unix()
	var currentCount int64

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
					break loopNext
				}
				sources = append(sources[:idx], sources[idx+1:]...)
				continue loopNext
			}
			if nextData == nil || cmpNextData.DateTime < nextData.DateTime ||
				(cmpNextData.Frequency < nextData.Frequency && cmpNextData.Frequency >= 0) {
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
		pendingBars = append(pendingBars, bar)
		currentCount++
		if len(pendingBars) >= dbBatchCreateSize {
			output.Create(pendingBars)
			pendingBars = []*BarData{}
		}
		if time.Now().Unix()-lastTime > 10 {
			lastTime = time.Now().Unix()
			mergeBar.Set64(currentCount)
		}
		sources[nextIdx].Next()
	}
	if len(pendingBars) > 0 {
		output.Create(pendingBars)
	}
	mergeBar.Finish()
	return nil
}
