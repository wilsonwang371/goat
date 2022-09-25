package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var SkippedBars = promauto.NewCounter(prometheus.CounterOpts{
	Name: "goat_skipped_bars",
	Help: "The total number of skipped bars",
})
