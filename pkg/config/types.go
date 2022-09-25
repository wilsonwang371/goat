package config

const (
	DebugLevel = iota + 1
	InfoLevel
	WarnLevel
	ErrorLevel
)

const (
	NotifyIsMobileFlag = 1 << iota
	NotifyIsEmailFlag
)

const DataFeedMaxPendingBars = 10000

type Config struct {
	KVDB   string `mapstructure:"kvdb"`
	Symbol string `mapstructure:"symbol"`
	Dump   struct {
		BarDumpDB     string `mapstructure:"bardumpdb"`       // name of db to dump live feed data, leave empty to disable
		RemoveOldBars bool   `mapstructure:"delete_old_bars"` // delete db if exist
	} `mapstructure:"dump"`
	Live struct {
		TradingView struct {
			User string `mapstructure:"user"`
			Pass string `mapstructure:"password"`
		} `mapstructure:"tradingview"`
	} `mapstructure:"live"`
	Notification struct {
		Twilio struct {
			Enabled bool     `mapstructure:"enabled"`
			Level   int      `mapstructure:"level"`
			SID     string   `mapstructure:"sid"`
			Token   string   `mapstructure:"token"`
			From    string   `mapstructure:"from"`
			To      []string `mapstructure:"to"`
		} `mapstructure:"twilio"`
		Pushover struct {
			Enabled bool     `mapstructure:"enabled"`
			Level   int      `mapstructure:"level"`
			Token   string   `mapstructure:"token"`
			Keys    []string `mapstructure:"keys"`
		} `mapstructure:"pushover"`
		Email struct {
			Enabled  bool     `mapstructure:"enabled"`
			Level    int      `mapstructure:"level"`
			Host     string   `mapstructure:"host"`
			Port     int      `mapstructure:"port"`
			From     string   `mapstructure:"from"`
			To       []string `mapstructure:"to"`
			User     string   `mapstructure:"user"`
			Password string   `mapstructure:"password"`
		} `mapstructure:"email"`
	} `mapstructure:"notification"`
}
