package fact

// HandlerSkipReason is an enumeration of the reasons a handler can be skipped
// while dispatching a message or ticking.
type HandlerSkipReason byte

const (
	// HandlerTypeDisabled indicates that a handler was skipped because all
	// handlers of that type have been disabled using an engine option.
	HandlerTypeDisabled HandlerSkipReason = 'T'

	// IndividualHandlerDisabled indicates that a handler was skipped because
	// that specific handler was disabled using an engine option.
	IndividualHandlerDisabled HandlerSkipReason = 'I'

	// IndividualHandlerDisabledByConfiguration indicates that a handler was
	// skipped because it was disabled by a call to Disable() in its Configure()
	// method.
	IndividualHandlerDisabledByConfiguration HandlerSkipReason = 'C'
)
