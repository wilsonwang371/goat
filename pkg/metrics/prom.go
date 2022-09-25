package metrics

import (
	"fmt"
	"net/http"

	"goat/pkg/common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var SkippedBars = promauto.NewCounter(prometheus.CounterOpts{
	Name: "goat_skipped_bars",
	Help: "The total number of skipped bars",
})

func StartMetricsServer() {
	go func() {
		// setup prometheus metrics
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", common.MetricsPort), nil)
	}()
}
