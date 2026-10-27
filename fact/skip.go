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

// DispatchCycleSkipReason is an enumeration of the reasons a dispatch cycle can
// be skipped.
type DispatchCycleSkipReason byte

const (
	// UnroutedMessageType indicates that a dispatch cycle was skipped because
	// the application does not have any routes for the message type.
	UnroutedMessageType DispatchCycleSkipReason = 'U'

	// OutboundOnlyMessageType indicates that a dispatch cycle was skipped
	// because the message type is produced by one or more handlers, but is not
	// handled by any of them.
	OutboundOnlyMessageType DispatchCycleSkipReason = 'O'
)
