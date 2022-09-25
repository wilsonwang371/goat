package db

import (
	"os"

	"goat/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const dataBatchSize = 1024 * 96

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

func NewSQLiteDataBase(dbpath string, removeOldData bool) *DB {
	if _, err := os.Stat(dbpath); err != nil && os.IsNotExist(err) {
		// file does not exist
		logger.Logger.Info("using new database file", zap.String("dbpath", dbpath))
	} else {
		// file exists
		if removeOldData {
			logger.Logger.Info("delete existing db", zap.String("dbpath", dbpath))
			err = os.Remove(dbpath)
			if err != nil {
				logger.Logger.Fatal("failed to remove db file", zap.Error(err))
				panic(err)
			}
		}
	}
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		logger.Logger.Error("failed to connect database", zap.Error(err))
		panic(err)
	}
	db.AutoMigrate(&BarData{})

	return &DB{
		db,
		make(chan *BarData, dataBatchSize),
		nil,
	}
}

func (db *DB) fetchAll() {
	var data []*BarData

	db.err = nil
	startIdx := 0
	for {
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

func (db *DB) FetchAll(bg bool) int64 {
	count := int64(0)
	db.Model(&BarData{}).Count(&count)
	if bg {
		go func() {
			db.fetchAll()
		}()
	} else {
		db.fetchAll()
	}
	return count
}

func (db *DB) Next() (*BarData, error) {
	data, ok := <-db.dataChan
	if !ok {
		return nil, db.err
	}
	return data, nil
}
