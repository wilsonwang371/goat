package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type BarData struct {
	gorm.Model
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

func NewSQLiteDataBase(dbpath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&BarData{})
	return db
}
