package fact

import (
	"time"

	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/envelope"
)

// DispatchCycleBegun indicates that Engine.Dispatch() has been called with a
// message that is able to be routed to at least one handler.
type DispatchCycleBegun struct {
	Envelope            *envelope.Envelope
	EngineTime          time.Time
	EnabledHandlerTypes map[config.HandlerType]bool
	EnabledHandlers     map[string]bool
}

// DispatchCycleCompleted indicates that a call Engine.Dispatch() has completed.
type DispatchCycleCompleted struct {
	Envelope            *envelope.Envelope
	Error               error
	EnabledHandlerTypes map[config.HandlerType]bool
	EnabledHandlers     map[string]bool
}

// DispatchBegun indicates that Engine.Dispatch() has been called with a
// message that is able to be routed to at least one handler.
type DispatchBegun struct {
	Envelope *envelope.Envelope
}

// DispatchCompleted indicates that a call Engine.Dispatch() has completed.
type DispatchCompleted struct {
	Envelope *envelope.Envelope
	Error    error
}

// HandlingBegun indicates that a message is about to be handled by a specific
// handler.
type HandlingBegun struct {
	Handler  config.Handler
	Envelope *envelope.Envelope
}

// HandlingCompleted indicates that a message has been handled by a specific
// handler, either successfully or unsuccessfully.
type HandlingCompleted struct {
	Handler  config.Handler
	Envelope *envelope.Envelope
	Error    error
}

// HandlingSkipped indicates that a message has been not been handled by a
// specific handler, either because all handlers of that type are or the handler
// itself is disabled.
type HandlingSkipped struct {
	Handler  config.Handler
	Envelope *envelope.Envelope
	Reason   HandlerSkipReason
}
