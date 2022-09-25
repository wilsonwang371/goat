package metrics

import (
	"fmt"
	"net/http"

	"goat/pkg/common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	BarsNotSaved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goat_bars_not_saved",
		Help: "The total number of bars not saved from recovery",
	})
	OutdatedBars = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goat_outdated_bars",
		Help: "The total number of outdated bars",
	})
)

func StartMetricsServer() {
	go func() {
		// setup prometheus metrics
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", common.MetricsPort), nil)
	}()
}
