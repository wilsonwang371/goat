package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"goat/pkg/config"
	"goat/pkg/core"
	"goat/pkg/feedgen"
	"goat/pkg/js"
	"goat/pkg/logger"
	"goat/pkg/notify"
	"goat/pkg/util"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	runScriptFile string
	runDataSource string

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run command executes the specified strategy script",
		Long: `run command executes the specified strategy script.
	`,
		Run: RunFunction,
	}
)

func RunFunction(cmd *cobra.Command, args []string) {
	// handle panic
	defer util.PanicHandler(notify.NewEmailNotifier(&cfg))

	logger.Logger.Debug("running script", zap.String("runScriptFile", runScriptFile))

	ctx := util.NewTerminationContext()

	// setup provider, data generator and feed
	gen := GetFeedGenerator()
	if gen == nil {
		logger.Logger.Error("failed to create feed generator")
		os.Exit(1)
	}

	feed := core.NewGenericDataFeed(ctx, &config.Config{}, gen, nil, 100, "")

	// setup js runtime
	rt := js.NewStrategyRuntime(ctx, &cfg, feed, nil)
	script, err := ioutil.ReadFile(runScriptFile)
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

func GetFeedGenerator() core.FeedGenerator {
	if u, err := url.ParseRequestURI(runDataSource); err != nil {
		ext := filepath.Ext(runDataSource)
		switch ext {
		case ".csv":
			return feedgen.NewCSVBarFeedGenerator(runDataSource, "symbol", core.UNKNOWN)
		default:
			logger.Logger.Error("unsupported file type", zap.String("fileType", ext))
			return nil
		}
	} else {
		switch u.Scheme {
		case "file":
			ext := filepath.Ext(u.Path)
			switch ext {
			case ".csv":
				return feedgen.NewCSVBarFeedGenerator(u.Path, "symbol", core.UNKNOWN)
			default:
				logger.Logger.Error("unsupported file type", zap.String("fileType", ext))
				return nil
			}
		case "remote":
			switch u.Host {
			case "yahoo":
				return feedgen.NewYahooBarFeedGenerator(cfg.Symbol, core.UNKNOWN)
			default:
				logger.Logger.Error("unsupported remote data source",
					zap.String("runDataSource", u.Host))
				return nil
			}
		default:
			logger.Logger.Error("unknown data source", zap.String("runDataSource", runDataSource))
			return nil
		}
	}
}

func init() {
	runCmd.PersistentFlags().StringVarP(&runScriptFile, "strategy", "f", "",
		"strategy js script file")
	runCmd.MarkPersistentFlagRequired("strategy")

	runCmd.PersistentFlags().StringVarP(&runDataSource, "datasource", "s", "",
		"data source(support url scheme: csv, yahoo) e.g. csv:///path/to/file.csv or yahoo://")
	runCmd.MarkPersistentFlagRequired("datasource")

	rootCmd.AddCommand(runCmd)
}
