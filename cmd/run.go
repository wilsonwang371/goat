package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"goalgotrade/pkg/core"
	"goalgotrade/pkg/feedgen"
	"goalgotrade/pkg/js"
	"goalgotrade/pkg/logger"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	scriptFile string
	dataType   string
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

	rt := js.NewRuntime(cfg.DB)
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
	if dataType == "csv" {
		return feedgen.NewCSVBarFeedGenerator(dataSource, "SYMBOL", core.UNKNOWN)
	} else {
		logger.Logger.Error("unknown data type", zap.String("dataType", dataType))
		os.Exit(1)
	}
	return nil
}

func init() {
	runCmd.PersistentFlags().StringVarP(&scriptFile, "strategy", "f", "",
		"strategy js script file")
	runCmd.MarkPersistentFlagRequired("strategy")

	runCmd.PersistentFlags().StringVarP(&dataType, "datatype", "t", "",
		"data type(csv)")
	runCmd.MarkPersistentFlagRequired("datatype")

	runCmd.PersistentFlags().StringVarP(&dataSource, "datasource", "s", "",
		"data source")
	runCmd.MarkPersistentFlagRequired("datasource")

	rootCmd.AddCommand(runCmd)
}
