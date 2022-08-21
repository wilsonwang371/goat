package cmd

import "github.com/spf13/cobra"

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
	// TODO: implement convert command
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
