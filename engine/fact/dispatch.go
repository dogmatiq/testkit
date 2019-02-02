package fact

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/enginekit/handler"
)

// DispatchCycleBegun indicates that Engine.Dispatch() has been called with a
// message that is able to be routed to at least one handler.
type DispatchCycleBegun struct {
	Envelope        *envelope.Envelope
	EngineTime      time.Time
	EnabledHandlers map[handler.Type]bool
}

// DispatchCycleCompleted indicates that a call Engine.Dispatch() has completed.
type DispatchCycleCompleted struct {
	Envelope        *envelope.Envelope
	Error           error
	EnabledHandlers map[handler.Type]bool
}

// DispatchCycleSkipped indicates that Engine.Dispatch() has been called
// with a message that is not routed to any handlers.
//
// Note that when dispatch is called with an unroutable message, it is unknown
// whether it was intended to be a command or an event.
type DispatchCycleSkipped struct {
	Message         dogma.Message
	EngineTime      time.Time
	EnabledHandlers map[handler.Type]bool
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
	HandlerName string
	HandlerType handler.Type
	Envelope    *envelope.Envelope
}

// HandlingCompleted indicates that a message has been handled by a specific
// handler, either successfully or unsucessfully.
type HandlingCompleted struct {
	HandlerName string
	HandlerType handler.Type
	Envelope    *envelope.Envelope
	Error       error
}

// HandlingSkipped indicates that a message has been not been handled by a
// specific handler, because handlers of that type are disabled.
type HandlingSkipped struct {
	HandlerName string
	HandlerType handler.Type
	Envelope    *envelope.Envelope
}
