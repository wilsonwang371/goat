package cmd

import (
	"github.com/spf13/cobra"
)

var (
	mergeDataSource1 string
	mergeDataSource2 string
	mergeOutput      string
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge command merges two data inputs into one",
	Long: `merge command merges two data inputs into one.
	`,
	Run: ConvertFunction,
}

func MergeFunction(cmd *cobra.Command, args []string) {
	// TODO: implement merge command
}

func init() {
	// TODO: add flags
	rootCmd.AddCommand(mergeCmd)
}
