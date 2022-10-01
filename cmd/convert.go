package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"goat/pkg/cmd/convert"
	"goat/pkg/db"
	"goat/pkg/js"
	"goat/pkg/logger"
	"goat/pkg/notify"
	"goat/pkg/util"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	convertScriptFile string
	convertDataSource string
	convertFileType   string
	convertOutput     string
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "convert command converts one data format to the one supported by goat",
	Long: `convert command converts one data format to the one supported by goat.
	`,
	Run: ConvertFunction,
}

func ConvertFunction(cmd *cobra.Command, args []string) {
	// handle panic
	defer util.PanicHandler(notify.NewEmailNotifier(&cfg))

	rt := js.NewDBConvertRuntime(&cfg)
	script, err := ioutil.ReadFile(convertScriptFile)
	if err != nil {
		logger.Logger.Error("failed to read script file", zap.Error(err),
			zap.String("script", convertScriptFile))
		os.Exit(1)
	}
	compiledScript, err := rt.Compile(string(script))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err := rt.Execute(compiledScript); err != nil {
		logger.Logger.Error("failed to execute script", zap.Error(err))
		os.Exit(1)
	}

	dbsource := convert.NewDBSource(convertDataSource, convertFileType)
	dboutput, err := db.NewSQLiteDataBase(convertOutput, true)
	if err != nil {
		logger.Logger.Error("failed to create output database", zap.Error(err))
		os.Exit(1)
	}
	if err := rt.Convert(dbsource, dboutput); err != nil {
		logger.Logger.Error("failed to convert data", zap.Error(err))
		os.Exit(1)
	}
}

func init() {
	convertCmd.PersistentFlags().StringVarP(&convertDataSource, "datasource", "s", "",
		"data source(support url scheme: csv, yahoo) e.g. csv:///path/to/file.csv or yahoo://")
	convertCmd.MarkPersistentFlagRequired("datasource")

	convertCmd.PersistentFlags().StringVarP(&convertOutput, "output-file", "o", "",
		"output file path")
	convertCmd.MarkPersistentFlagRequired("output-file")

	convertCmd.PersistentFlags().StringVarP(&convertScriptFile, "script", "f", "",
		"source data column mapping js file")
	convertCmd.MarkPersistentFlagRequired("script")

	convertCmd.PersistentFlags().StringVarP(&convertFileType, "type", "t", "",
		"source data file type(sqlite)")
	convertCmd.MarkPersistentFlagRequired("type")

	rootCmd.AddCommand(convertCmd)
}
