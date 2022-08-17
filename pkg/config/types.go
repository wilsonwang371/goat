package config

type Config struct {
	KVDB      string `mapstructure:"kvdb"`
	Symbol    string `mapstructure:"symbol"`
	BarDumpDB string `mapstructure:"bardumpdb"` // name of db to dump live feed data, leave empty to disable
	Live      struct {
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
