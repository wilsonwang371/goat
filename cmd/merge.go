package cmd

import (
	"os"
	"strings"

	"goat/pkg/db"
	"goat/pkg/logger"
	"goat/pkg/notify"
	"goat/pkg/util"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	mergeDataSourceList string
	mergeOutput         string
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge command merges two dumpdb data inputs into one",
	Long: `merge command merges two dumpdb data inputs into one.
	`,
	Run: MergeFunction,
}

func MergeFunction(cmd *cobra.Command, args []string) {
	// handle panic
	defer util.PanicHandler(notify.NewEmailNotifier(&cfg))

	var sources []*db.DB
	sourceNames := strings.Split(mergeDataSourceList, ",")
	for _, name := range sourceNames {
		tmp, err := db.NewSQLiteDataBase(name, false)
		if err != nil {
			logger.Logger.Error("failed to open database", zap.Error(err))
			os.Exit(1)
		}
		sources = append(sources, tmp)
	}

	output, err := db.NewSQLiteDataBase(mergeOutput, true)
	if err != nil {
		logger.Logger.Error("failed to create output database", zap.Error(err))
		os.Exit(1)
	}

	if err := db.MergeDBs(output, sources); err != nil {
		logger.Logger.Error("failed to merge databases", zap.Error(err))
		os.Exit(1)
	}
}

func init() {
	mergeCmd.PersistentFlags().StringVarP(&mergeDataSourceList,
		"datasources", "s", "", "data source list, separated by comma")
	mergeCmd.MarkPersistentFlagRequired("datasources")

	mergeCmd.PersistentFlags().StringVarP(&mergeOutput, "output-file", "o", "",
		"output file path")
	mergeCmd.MarkPersistentFlagRequired("output-file")

	rootCmd.AddCommand(mergeCmd)
}
