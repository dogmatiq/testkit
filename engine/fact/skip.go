package fact

// HandlerSkipReason is an enumeration of the reasons a handler can be skipped
// while dispatching a message or ticking.
type HandlerSkipReason byte

const (
	// HandlerTypeDisabled indicates that a handler skipped because all handlers
	// of that type have been disabled.
	HandlerTypeDisabled HandlerSkipReason = 'T'

	// IndividualHandlerDisabled indicates that a handler was skipped because
	// that specific handler was disabled.
	IndividualHandlerDisabled HandlerSkipReason = 'I'
)
