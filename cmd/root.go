package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"honnef.co/go/tools/config"
)

var rootCmd = &cobra.Command{
	Use:   "goalgotrade",
	Short: "goalgotrade is a tool for trading",
	Long: `goalgotrade is a tool for trading.

It is a tool for trading.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("goalgotrade is a tool for trading")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goalgotrade.yaml)")
}

var (
	cfgFile string
	cfg     config.Config
)

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".goalgotrade")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME")
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
