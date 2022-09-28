package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"goat/pkg/core"
	"goat/pkg/feedgen"
	"goat/pkg/js"
	"goat/pkg/logger"
	"goat/pkg/metrics"
	"goat/pkg/notify"
	"goat/pkg/util"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	liveScriptFile     string
	liveRecoveryDBFile string

	feedProviders string
	runWg         *sync.WaitGroup

	liveCmd = &cobra.Command{
		Use:   "live",
		Short: "live command executes the specified strategy script using live data",
		Long: `live command executes the specified strategy script using live data.
	`,
		Run: runLiveCmd,
	}
)

// startLive function is used by js runtime to start live feed
func startLive() error {
	logger.Logger.Info("start live strategy data feed")
	if runWg == nil {
		panic("runWg is nil")
	}
	// unblock bar feed generator
	runWg.Done()
	return nil
}

func runLiveCmd(cmd *cobra.Command, args []string) {
	// handle panic
	defer util.PanicHandler(notify.NewEmailNotifier(&cfg))

	logger.Logger.Debug("running script", zap.String("liveScriptFile", liveScriptFile))
	logger.Logger.Debug("running with symbol", zap.String("symbol", cfg.Symbol))

	ctx := util.NewTerminationContext()

	// setup provider, data generator and feed
	providers := strings.Split(feedProviders, ",")
	gen, wg := GetLiveFeedGenerator(ctx, providers)
	if gen == nil {
		logger.Logger.Error("failed to create feed generator")
		os.Exit(1)
	}
	runWg = wg
	feed := core.NewGenericDataFeed(ctx, &cfg, gen, nil, 250, liveRecoveryDBFile)

	// setup metrics server
	metrics.StartMetricsServer()

	// setup js runtime
	rt := js.NewStrategyRuntime(ctx, &cfg, feed, startLive)
	script, err := ioutil.ReadFile(liveScriptFile)
	if err != nil {
		logger.Logger.Error("failed to read script file", zap.Error(err))
		os.Exit(1)
	}
	if compiledScript, err := rt.Compile(string(script)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		// starting from here, we start to run the strategy
		if _, err := rt.Execute(compiledScript); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		sel := js.NewJSStrategyEventListener(rt)
		broker := core.NewDummyBroker(feed)
		strategy := core.NewStrategyController(ctx, &cfg, sel, broker, feed)

		strategy.Run()
	}
}

func CreateOneProvider(p string) (feedgen.BarDataProvider, error) {
	var provider feedgen.BarDataProvider
	if strings.EqualFold(p, "fake") {
		provider = feedgen.NewFakeDataProvider()
	} else if strings.EqualFold(p, "tradingview") {
		provider = feedgen.NewTradingViewDataProvider(cfg.Live.TradingView.User,
			cfg.Live.TradingView.Pass)
	} else if strings.EqualFold(p, "fx678") {
		provider = feedgen.NewFx678DataProvider()
	} else if strings.EqualFold(p, "goldpriceorg") {
		provider = feedgen.NewGoldPriceOrgDataProvider()
	} else {
		logger.Logger.Error("unknown live feed provider", zap.String("provider", p))
		return nil, fmt.Errorf("unknown live feed provider: %s", p)
	}
	return provider, nil
}

func GetLiveFeedGenerator(ctx context.Context, providers []string) (core.FeedGenerator, *sync.WaitGroup) {
	if len(providers) == 0 {
		logger.Logger.Error("no feed provider specified")
		os.Exit(1)
	} else if len(providers) == 1 {
		p, err := CreateOneProvider(providers[0])
		if err != nil {
			logger.Logger.Error("failed to create feed provider", zap.Error(err))
			os.Exit(1)
		}
		gen := feedgen.NewLiveBarFeedGenerator(ctx,
			p,
			cfg.Symbol,
			[]core.Frequency{core.REALTIME},
			100)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go gen.WaitAndRun(wg)
		return gen, wg
	} else {
		pArr := make([]feedgen.BarDataProvider, len(providers))
		for i, pStr := range providers {
			p, err := CreateOneProvider(pStr)
			if err != nil {
				logger.Logger.Error("failed to create feed provider", zap.Error(err))
				os.Exit(1)
			}
			pArr[i] = p
		}
		gen := feedgen.NewMultiLiveBarFeedGenerator(ctx,
			pArr,
			cfg.Symbol,
			[]core.Frequency{core.REALTIME},
			100)

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go gen.WaitAndRun(wg)
		return gen, wg
	}
	return nil, nil // should not reach here
}

func init() {
	liveCmd.PersistentFlags().StringVarP(&liveScriptFile, "strategy", "f", "",
		"strategy js script file")
	liveCmd.MarkPersistentFlagRequired("strategy")
	liveCmd.PersistentFlags().StringVarP(&feedProviders, "providers", "p", "", "live feed data providers name, separated by comma")

	liveCmd.PersistentFlags().StringVarP(&liveRecoveryDBFile, "recovery-db", "r", "",
		"goat db file that will be replayed before go live")

	rootCmd.AddCommand(liveCmd)
}
