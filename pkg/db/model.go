package db

import (
	"os"

	"goat/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const dataBatchSize = 4096 * 2

type BarData struct {
	gorm.Model
	Symbol    string  `json:"symbol"`
	DateTime  int64   `json:"dateTime"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
	AdjClose  float64 `json:"adjClose"`
	Frequency int64   `json:"frequency"`
	Note      string  `json:"note"`
}

type DB struct {
	*gorm.DB
	dataChan chan *BarData
	err      error
}

func NewSQLiteDataBase(dbpath string) *DB {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		logger.Logger.Error("failed to connect database", zap.Error(err))
		os.Exit(1)
	}
	db.AutoMigrate(&BarData{})

	return &DB{
		db,
		make(chan *BarData, dataBatchSize),
		nil,
	}
}

func (db *DB) fetchAll() {
	db.err = nil
	startIdx := 0
	for {
		var data []*BarData
		if err := db.Model(&BarData{}).Order("id").Offset(startIdx).Limit(dataBatchSize).Find(&data).Error; err != nil {
			logger.Logger.Error("failed to fetch data", zap.Error(err))
			db.err = err
			break
		}
		if len(data) == 0 {
			db.err = nil
			break
		}
		for _, d := range data {
			db.dataChan <- d
		}
		startIdx += dataBatchSize
	}
	close(db.dataChan)
}

func (db *DB) FetchAll(bg bool) {
	if bg {
		go func() {
			db.fetchAll()
		}()
	} else {
		db.fetchAll()
	}
}

func (db *DB) Next() (*BarData, error) {
	data, ok := <-db.dataChan
	if !ok {
		return nil, db.err
	}
	return data, nil
}
