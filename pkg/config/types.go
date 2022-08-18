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
		Pushover struct {
			Token string   `mapstructure:"token"`
			Keys  []string `mapstructure:"keys"`
		} `mapstructure:"pushover"`
		Email struct {
			Host     string   `mapstructure:"host"`
			Port     int      `mapstructure:"port"`
			From     string   `mapstructure:"from"`
			To       []string `mapstructure:"to"`
			User     string   `mapstructure:"user"`
			Password string   `mapstructure:"password"`
		} `mapstructure:"email"`
	} `mapstructure:"notification"`
}
