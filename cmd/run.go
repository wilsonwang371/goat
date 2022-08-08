package cmd

import (
	"fmt"
	"goalgotrade/pkg/js"
	"goalgotrade/pkg/logger"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	scriptFile string
	runCmd     = &cobra.Command{
		Use:   "run",
		Short: "run command executes the specified script",
		Long: `run command executes the specified script.
	`,
		Run: func(cmd *cobra.Command, args []string) {
			rt := js.NewRuntime()
			// logger.Logger.Info("scriptFile:", zap.String("scriptFile", scriptFile))
			if script, err := ioutil.ReadFile(scriptFile); err != nil {
				logger.Logger.Error("failed to read script file", zap.Error(err))
				os.Exit(1)
			} else {
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
				}
			}
		},
	}
)

func init() {
	runCmd.PersistentFlags().StringVarP(&scriptFile, "script", "s", "", "strategy js script file")
	runCmd.MarkPersistentFlagRequired("script")
	rootCmd.AddCommand(runCmd)
}
