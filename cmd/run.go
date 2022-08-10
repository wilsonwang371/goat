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
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	scriptFile       string
	runningDuration  time.Duration
	live             bool
	liveFeedProvider string

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run command executes the specified script",
		Long: `run command executes the specified script.
	`,
		Run: RunFunction,
	}
)

func RunFunction(cmd *cobra.Command, args []string) {
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

		var gen core.FeedGenerator
		if live {
			gen = GetLiveFeedGenerator()
		} else {
			// TODO: other data generator
		}

		if gen == nil {
			logger.Logger.Error("failed to create feed generator")
			os.Exit(1)
		}

		feed := core.NewGenericDataFeed(gen, 100)

		sel := js.NewJSStrategyEventListener(rt)
		broker := core.NewDummyBroker(feed)
		strategy := core.NewStrategyController(sel, broker, feed)

		if runningDuration > 0 {
			go strategy.Run()
			select {
			case <-time.After(runningDuration):
				strategy.Stop()
				logger.Logger.Debug("strategy stopped")
			}
		} else {
			strategy.Run()
		}
	}
}

func GetLiveFeedGenerator() core.FeedGenerator {
	var provider feedgen.BarDataProvider
	if strings.EqualFold(liveFeedProvider, "fake") {
		provider = feedgen.NewFakeDataProvider()
	} else {
		logger.Logger.Error("unknown live feed provider", zap.String("provider", liveFeedProvider))
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
	runCmd.PersistentFlags().StringVarP(&scriptFile, "script", "s", "",
		"strategy js script file")
	runCmd.MarkPersistentFlagRequired("script")
	runCmd.PersistentFlags().DurationVarP(&runningDuration, "time", "t", time.Duration(0),
		"strategy maximum running duration. (default: no limit)")
	runCmd.PersistentFlags().BoolVarP(&live, "live", "l", false, "use live data")
	runCmd.PersistentFlags().StringVarP(&liveFeedProvider, "provider", "p", "", "live feed data provider name")
	runCmd.MarkFlagsRequiredTogether("live", "provider")
	rootCmd.AddCommand(runCmd)
}
