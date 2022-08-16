package core

// Frequency ...
type Frequency int64

// INVALID ...
const (
	// Enum like class for bar frequencies. Valid values are:

	// * **Frequency.TRADE**: The bar represents a single trade.
	// * **Frequency.SECOND**: The bar summarizes the trading activity during 1 second.
	// * **Frequency.MINUTE**: The bar summarizes the trading activity during 1 minute.
	// * **Frequency.HOUR**: The bar summarizes the trading activity during 1 hour.
	// * **Frequency.DAY**: The bar summarizes the trading activity during 1 day.
	// * **Frequency.WEEK**: The bar summarizes the trading activity during 1 week.
	// * **Frequency.MONTH**: The bar summarizes the trading activity during 1 month.
	// * **Frequency.YEAR**: The bar summarizes the trading activity during 1 year.

	// It is important for frequency values to get bigger for bigger windows.
	UNKNOWN  Frequency = -9999
	INVALID  Frequency = -999
	RESET    Frequency = -99
	TRADE    Frequency = -1
	REALTIME Frequency = 0
	SECOND   Frequency = 1
	MINUTE   Frequency = 60
	HOUR     Frequency = 60 * 60
	HOUR_4   Frequency = 60 * 60 * 4
	DAY      Frequency = 24 * 60 * 60
	WEEK     Frequency = 24 * 60 * 60 * 7
	MONTH    Frequency = 24 * 60 * 60 * 31
	YEAR     Frequency = 24 * 60 * 60 * 365
)
