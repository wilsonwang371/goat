package common

type Frequency int64

const (
	// Enum like class for bar frequencies. Valid values are:

	// * **Frequency.TRADE**: The bar represents a single trade.
	// * **Frequency.SECOND**: The bar summarizes the trading activity during 1 second.
	// * **Frequency.MINUTE**: The bar summarizes the trading activity during 1 minute.
	// * **Frequency.HOUR**: The bar summarizes the trading activity during 1 hour.
	// * **Frequency.DAY**: The bar summarizes the trading activity during 1 day.
	// * **Frequency.WEEK**: The bar summarizes the trading activity during 1 week.
	// * **Frequency.MONTH**: The bar summarizes the trading activity during 1 month.

	// It is important for frequency values to get bigger for bigger windows.
	Frequency_INVALID  Frequency = -999
	Frequency_RESET    Frequency = -99
	Frequency_TRADE    Frequency = -1
	Frequency_REALTIME Frequency = 0
	Frequency_SECOND   Frequency = 1
	Frequency_MINUTE   Frequency = 60
	Frequency_HOUR     Frequency = 60 * 60
	Frequency_HOUR_4   Frequency = 60 * 60 * 4
	Frequency_DAY      Frequency = 24 * 60 * 60
	Frequency_WEEK     Frequency = 24 * 60 * 60 * 7
	Frequency_MONTH    Frequency = 24 * 60 * 60 * 31
)

const (
	OrderState_UNKNOWN          OrderState = iota
	OrderState_INITIAL                     // Initial state.
	OrderState_SUBMITTED                   // Order has been submitted.
	OrderState_ACCEPTED                    // Order has been acknowledged by the broker.
	OrderState_CANCELED                    // Order has been canceled.
	OrderState_PARTIALLY_FILLED            // Order has been partially filled.
	OrderState_FILLED                      // Order has been completely filled.
)

const (
	OrderEventType_SUBMITTED        OrderEventType = iota + 1 // Order has been submitted.
	OrderEventType_ACCEPTED                                   // Order has been acknowledged by the broker.
	OrderEventType_CANCELED                                   // Order has been canceled.
	OrderEventType_PARTIALLY_FILLED                           // Order has been partially filled.
	OrderEventType_FILLED                                     // Order has been completely filled.
)
