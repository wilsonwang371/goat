package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "goalgotrade",
	Short: "goalgotrade is a tool for trading",
	Long: `goalgotrade is a tool for trading.

It is a tool for trading.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.goalgotrade.yaml)")
	rootCmd.PersistentFlags().StringVarP(&dbFile, "strategy-db", "d", "", "state db file used for strategy (default is using in-memory db)")
}

type Config struct {
	Live struct {
		TradingView struct {
			User string `mapstructure:"user"`
			Pass string `mapstructure:"password"`
		} `mapstructure:"tradingview"`
	} `mapstructure:"live"`
	Notification struct {
		Twilio struct {
			Sid   string `mapstructure:"sid"`
			Token string `mapstructure:"token"`
		} `mapstructure:"twilio"`
		PushOver struct {
			Key   string `mapstructure:"key"`
			Token string `mapstructure:"token"`
		} `mapstructure:"pushover"`
	} `mapstructure:"notification"`
}

var (
	cfgFile string
	dbFile  string
	cfg     Config
)

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".goalgotrade")
		viper.SetConfigType("json")
		viper.AddConfigPath("$HOME")
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("no config file found")
		return
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
