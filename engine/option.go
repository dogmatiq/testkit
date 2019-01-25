package engine

import (
	"time"

	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/handler"
)

// Option applies optional settings to an engine.
type Option func(*Engine) error

// WithResetter returns an engine option that registers a reset hook with the
// engine.
//
// fn is a function to be called whenever the engine is reset.
func WithResetter(fn func()) Option {
	if fn == nil {
		panic("fn must not be nil")
	}

	return func(e *Engine) error {
		e.resetters = append(e.resetters, fn)
		return nil
	}
}

// DispatchOption applies optional settings while dispatching a message.
type DispatchOption func(*dispatchOptions) error

// WithObserver returns a dispatch option that registers the given observer
// while the message is being dispatched.
//
// Multiple observers can be registered during a single dispatch.
func WithObserver(o fact.Observer) DispatchOption {
	if o == nil {
		panic("observer must not be nil")
	}

	return func(do *dispatchOptions) error {
		do.observers = append(do.observers, o)
		return nil
	}
}

// EnableHandlerType returns a dispatch option that enables or disables handlers
// of the given type.
//
// All handler types are enabled by default.
func EnableHandlerType(t HandlerType, enable bool) DispatchOption {
	t.MustValidate()

	return func(do *dispatchOptions) error {
		do.enabledHandlers[t] = enable
		return nil
	}
}

// WithCurrentTime returns a dispatch option that sets the engine's current time.
func WithCurrentTime(t time.Time) DispatchOption {
	return func(do *dispatchOptions) error {
		do.now = t
		return nil
	}
}

// dispatchOptions is a container for the options set via DispatchOption values.
type dispatchOptions struct {
	now             time.Time
	observers       fact.ObserverGroup
	enabledHandlers map[handler.Type]bool
}

// newDispatchOptions returns a new dispatchOptions with the given options.
func newDispatchOptions(options []DispatchOption) (*dispatchOptions, error) {
	do := &dispatchOptions{
		now: time.Now(),
		enabledHandlers: map[handler.Type]bool{
			AggregateHandlerType:   true,
			ProcessHandlerType:     true,
			IntegrationHandlerType: true,
			ProjectionHandlerType:  true,
		},
	}

	for _, opt := range options {
		if err := opt(do); err != nil {
			return nil, err
		}
	}

	return do, nil
}

// HandlerType is an enumeration of the types of messages handlers.
type HandlerType = handler.Type

const (
	// AggregateHandlerType is the handler type for dogma.AggregateMessageHandler.
	AggregateHandlerType = handler.AggregateType

	// ProcessHandlerType is the handler type for dogma.ProcessMessageHandler.
	ProcessHandlerType = handler.ProcessType

	// IntegrationHandlerType is the handler type for dogma.IntegrationMessageHandler.
	IntegrationHandlerType = handler.IntegrationType

	// ProjectionHandlerType is the handler type for dogma.ProjectionMessageHandler.
	ProjectionHandlerType = handler.ProjectionType
)
