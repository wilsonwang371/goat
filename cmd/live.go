package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"goat/pkg/core"
	"goat/pkg/feedgen"
	"goat/pkg/js"
	"goat/pkg/logger"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
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
	logger.Logger.Debug("running with symbol", zap.String("symbol", cfg.Symbol))

	rt := js.NewRuntime(&cfg, startLive)
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
		strategy := core.NewStrategyController(&cfg, sel, broker, feed)

		if val, err := rt.Execute(compiledScript); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println(val)
		}

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

func GetLiveFeedGenerator() (core.FeedGenerator, *sync.WaitGroup) {
	var gen *feedgen.LiveBarFeedGenerator
	providers := strings.Split(feedProviders, ",")

	if len(providers) == 0 {
		logger.Logger.Error("no feed provider specified")
		os.Exit(1)
	} else if len(providers) == 1 {
		provider, err := CreateOneProvider(providers[0])
		if err != nil {
			logger.Logger.Error("failed to create feed provider", zap.Error(err))
			os.Exit(1)
		}
		gen = feedgen.NewLiveBarFeedGenerator(
			provider,
			cfg.Symbol,
			[]core.Frequency{core.REALTIME},
			100)
	} else {
		// TODO: support multiple providers
		logger.Logger.Error("multiple feed providers not supported yet")
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go gen.WaitAndRun(wg)
	return gen, wg
}

func init() {
	liveCmd.PersistentFlags().StringVarP(&scriptFile, "strategy", "f", "",
		"strategy js script file")
	liveCmd.MarkPersistentFlagRequired("strategy")
	liveCmd.PersistentFlags().StringVarP(&feedProviders, "providers", "p", "", "live feed data providers name, separated by comma")
	rootCmd.AddCommand(liveCmd)
}
