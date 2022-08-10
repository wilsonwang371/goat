package cmd

import (
	"fmt"
	"goalgotrade/pkg/logger"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goalgotrade.yaml)")
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
		logger.Logger.Error("failed to read config file", zap.Error(err))
		return
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
