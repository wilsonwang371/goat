package cmd

import (
	"fmt"
	"goalgotrade/pkg/core"
	"goalgotrade/pkg/feedgen"
	"goalgotrade/pkg/js"
	"goalgotrade/pkg/logger"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	feedProvider string

	liveCmd = &cobra.Command{
		Use:   "live",
		Short: "live command executes the specified strategy script using live data",
		Long: `live command executes the specified strategy script using live data.
	`,
		Run: runLiveCmd,
	}
)

func runLiveCmd(cmd *cobra.Command, args []string) {
	logger.Logger.Debug("running script", zap.String("scriptFile", scriptFile))

	rt := js.NewRuntime()
	script, err := ioutil.ReadFile(scriptFile)
	if err != nil {
		logger.Logger.Error("failed to read script file", zap.Error(err))
		os.Exit(1)
	}
	if compiledScript, err := rt.Compile(string(script)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		if val, err := rt.Execute(compiledScript); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println(val)
		}

		gen := GetLiveFeedGenerator()
		if gen == nil {
			logger.Logger.Error("failed to create feed generator")
			os.Exit(1)
		}

		feed := core.NewGenericDataFeed(gen, 100)

		sel := js.NewJSStrategyEventListener(rt)
		broker := core.NewDummyBroker(feed)
		strategy := core.NewStrategyController(sel, broker, feed)

		strategy.Run()
	}
}

func GetLiveFeedGenerator() core.FeedGenerator {
	var provider feedgen.BarDataProvider
	if strings.EqualFold(feedProvider, "fake") {
		provider = feedgen.NewFakeDataProvider()
	} else {
		logger.Logger.Error("unknown live feed provider", zap.String("provider", feedProvider))
		os.Exit(1)
	}
	gen := feedgen.NewLiveBarFeedGenerator(
		provider,
		"XAUUSD",
		[]core.Frequency{core.REALTIME, core.DAY},
		100)
	go gen.Run()
	return gen
}

func init() {
	liveCmd.PersistentFlags().StringVarP(&scriptFile, "strategy", "f", "",
		"strategy js script file")
	liveCmd.MarkPersistentFlagRequired("strategy")
	liveCmd.PersistentFlags().StringVarP(&feedProvider, "provider", "p", "", "live feed data provider name")
	rootCmd.AddCommand(liveCmd)
}
