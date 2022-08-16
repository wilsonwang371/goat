package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"goalgotrade/pkg/core"
	"goalgotrade/pkg/feedgen"
	"goalgotrade/pkg/js"
	"goalgotrade/pkg/logger"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	feedProvider string
	runWg        *sync.WaitGroup

	liveCmd = &cobra.Command{
		Use:   "live",
		Short: "live command executes the specified strategy script using live data",
		Long: `live command executes the specified strategy script using live data.
	`,
		Run: runLiveCmd,
	}
)

func startLive() error {
	logger.Logger.Info("start live strategy data feed")
	if runWg == nil {
		panic("runWg is nil")
	}
	runWg.Done()
	return nil
}

func runLiveCmd(cmd *cobra.Command, args []string) {
	logger.Logger.Debug("running script", zap.String("scriptFile", scriptFile))
	logger.Logger.Debug("running with symbol", zap.String("symbol", cfg.Live.Symbol))

	rt := js.NewRuntime(cfg.DB, startLive)
	script, err := ioutil.ReadFile(scriptFile)
	if err != nil {
		logger.Logger.Error("failed to read script file", zap.Error(err))
		os.Exit(1)
	}
	if compiledScript, err := rt.Compile(string(script)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		gen, wg := GetLiveFeedGenerator()
		if gen == nil {
			logger.Logger.Error("failed to create feed generator")
			os.Exit(1)
		}
		runWg = wg

		feed := core.NewGenericDataFeed(gen, 100)
		sel := js.NewJSStrategyEventListener(rt)
		broker := core.NewDummyBroker(feed)
		strategy := core.NewStrategyController(sel, broker, feed)

		if val, err := rt.Execute(compiledScript); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println(val)
		}

		strategy.Run()
	}
}

func GetLiveFeedGenerator() (core.FeedGenerator, *sync.WaitGroup) {
	var provider feedgen.BarDataProvider
	if strings.EqualFold(feedProvider, "fake") {
		provider = feedgen.NewFakeDataProvider()
	} else if strings.EqualFold(feedProvider, "tradingview") {
		provider = feedgen.NewTradingViewDataProvider(cfg.Live.TradingView.User,
			cfg.Live.TradingView.Pass)
	} else {
		logger.Logger.Error("unknown live feed provider", zap.String("provider", feedProvider))
		os.Exit(1)
	}
	gen := feedgen.NewLiveBarFeedGenerator(
		provider,
		cfg.Live.Symbol,
		[]core.Frequency{core.REALTIME, core.DAY},
		100)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go gen.DeferredRun(wg)

	return gen, wg
}

func init() {
	liveCmd.PersistentFlags().StringVarP(&scriptFile, "strategy", "f", "",
		"strategy js script file")
	liveCmd.MarkPersistentFlagRequired("strategy")
	liveCmd.PersistentFlags().StringVarP(&feedProvider, "provider", "p", "", "live feed data provider name")
	rootCmd.AddCommand(liveCmd)
}
