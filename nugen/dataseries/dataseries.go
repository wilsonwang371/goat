package dataseries

import (
	"time"
)

// DataSeries ...
type DataSeries interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) error
	Times() []*time.Time
	AtIndex(index int) interface{}
	Len() int
}
