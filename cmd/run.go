package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"goalgotrade/pkg/core"
	"goalgotrade/pkg/feedgen"
	"goalgotrade/pkg/js"
	"goalgotrade/pkg/logger"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	scriptFile string
	dataSource string

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run command executes the specified strategy script",
		Long: `run command executes the specified strategy script.
	`,
		Run: RunFunction,
	}
)

func RunFunction(cmd *cobra.Command, args []string) {
	logger.Logger.Debug("running script", zap.String("scriptFile", scriptFile))

	rt := js.NewRuntime(cfg.DB, nil)
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

		gen := GetFeedGenerator()
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

func GetFeedGenerator() core.FeedGenerator {
	if u, err := url.ParseRequestURI(dataSource); err != nil {
		ext := filepath.Ext(dataSource)
		switch ext {
		case ".csv":
			return feedgen.NewCSVBarFeedGenerator(dataSource, "symbol", core.UNKNOWN)
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
					zap.String("dataSource", u.Host))
				return nil
			}
		default:
			logger.Logger.Error("unknown data source", zap.String("dataSource", dataSource))
			return nil
		}
	}
}

func init() {
	runCmd.PersistentFlags().StringVarP(&scriptFile, "strategy", "f", "",
		"strategy js script file")
	runCmd.MarkPersistentFlagRequired("strategy")

	runCmd.PersistentFlags().StringVarP(&dataSource, "datasource", "s", "",
		"data source(support url scheme: csv, yahoo) e.g. csv:///path/to/file.csv or yahoo://")
	runCmd.MarkPersistentFlagRequired("datasource")

	rootCmd.AddCommand(runCmd)
}
