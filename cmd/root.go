package cmd

import (
	"fmt"
	"os"

	"goat/pkg/profile"

	"goat/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "goat",
	Short: "goat is a tool for trading",
	Long: `goat is a tool for trading.

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
	profile.StartProfilingServer()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		"config file (default is $HOME/.goat.json)")
	rootCmd.PersistentFlags().StringVarP(&cfg.KVDB, "kv-db", "d", "",
		"state key-value db file used for strategy (default is using in-memory db)")
	rootCmd.PersistentFlags().StringVarP(&cfg.Dump.BarDumpDB, "bar-dump", "b", "",
		"sqlite file used for bar dump in execution (leave empty to disable bar dump during execution)")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Dump.DeleteIfExist, "delete-old-bar-dump", "D", false,
		"delete old bar dump file if it exists")
	rootCmd.PersistentFlags().StringVarP(&cfg.Symbol, "symbol", "S", "",
		"live feed data symbol name")
}

var (
	cfgFile string
	cfg     config.Config
)

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".goat")
		viper.SetConfigType("json")
		viper.AddConfigPath("$HOME")
	}
	viper.AutomaticEnv()
	viper.BindPFlag("symbol", rootCmd.PersistentFlags().Lookup("symbol"))

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("no config file found")
		return
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
