package fact

import (
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/engine/envelope"
)

// DispatchCycleBegun indicates that Engine.Dispatch() has been called with a
// message that is able to be routed to at least one handler.
type DispatchCycleBegun struct {
	Envelope            *envelope.Envelope
	EngineTime          time.Time
	EnabledHandlerTypes map[configkit.HandlerType]bool
}

// DispatchCycleCompleted indicates that a call Engine.Dispatch() has completed.
type DispatchCycleCompleted struct {
	Envelope            *envelope.Envelope
	Error               error
	EnabledHandlerTypes map[configkit.HandlerType]bool
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
	Handler  configkit.RichHandler
	Envelope *envelope.Envelope
}

// HandlingCompleted indicates that a message has been handled by a specific
// handler, either successfully or unsuccessfully.
type HandlingCompleted struct {
	Handler  configkit.RichHandler
	Envelope *envelope.Envelope
	Error    error
}

// HandlingSkipped indicates that a message has been not been handled by a
// specific handler, because handlers of that type are disabled.
type HandlingSkipped struct {
	Handler  configkit.RichHandler
	Envelope *envelope.Envelope
}
