package order

// OrderState ...
type OrderState int

// UNKNOWN ...
const (
	OrderStateUNKNOWN          OrderState = iota
	OrderStateINITIAL                     // Initial state.
	OrderStateSUBMITTED                   // Order has been submitted.
	OrderStateACCEPTED                    // Order has been acknowledged by the broker.
	OrderStateCANCELED                    // Order has been canceled.
	OrderStatePARTIALLY_FILLED            // Order has been partially filled.
	OrderStateFILLED                      // Order has been completely filled.
)

// OrderEventType ...
type OrderEventType int

// OrderEventSUBMITTED ...
const (
	OrderEventSUBMITTED        OrderEventType = iota + 1 // Order has been submitted.
	OrderEventACCEPTED                                   // Order has been acknowledged by the broker.
	OrderEventCANCELED                                   // Order has been canceled.
	OrderEventPARTIALLY_FILLED                           // Order has been partially filled.
	OrderEventFILLED                                     // Order has been completely filled.
)
