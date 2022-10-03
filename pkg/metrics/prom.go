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
		Help: "The total number of bars read but not saved in dump while replaying recovery db",
	})
	OutdatedDBBars = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goat_outdated_db_bars",
		Help: "The total number of outdated bars in recovery db",
	})
	OutOfOrderBars = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goat_out_of_order_bars",
		Help: "The total number of out of order bars",
	})
	OnBarsCalledCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goat_on_bars_called_count",
		Help: "The total number of onBars() called",
	})
	OnIdleCalledCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goat_on_idle_called_count",
		Help: "The total number of onIdle() called",
	})
)

func StartMetricsServer() {
	go func() {
		// setup prometheus metrics
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", common.MetricsPort), nil)
	}()
}
