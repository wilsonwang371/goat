package db

import (
	"os"

	"goat/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
}

func NewSQLiteDataBase(dbpath string) *DB {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		logger.Logger.Error("failed to connect database", zap.Error(err))
		os.Exit(1)
	}
	db.AutoMigrate(&BarData{})

	return &DB{db}
}

func (db *DB) IterateRows(f func(*BarData)) error {
	if rows, err := db.Model(&BarData{}).Order("primarykey").Rows(); err != nil {
		return err
	} else {
		defer rows.Close()
		for rows.Next() {
			data := &BarData{}
			if err := db.ScanRows(rows, data); err != nil {
				return err
			}
			f(data)
		}
	}
	return nil
}
