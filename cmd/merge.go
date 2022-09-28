package cmd

import (
	"goat/pkg/notify"
	"goat/pkg/util"

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
	Run: MergeFunction,
}

func MergeFunction(cmd *cobra.Command, args []string) {
	// handle panic
	defer util.PanicHandler(notify.NewEmailNotifier(&cfg))

	// TODO: implement merge command
}

func init() {
	// TODO: add flags
	rootCmd.AddCommand(mergeCmd)
}
