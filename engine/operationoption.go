package engine

import (
	"time"

	"github.com/dogmatiq/enginekit/handler"

	"github.com/dogmatiq/testkit/engine/fact"
)

// OperationOption applies optional settings while dispatching a message or
// performing a tick.
type OperationOption func(*operationOptions)

// WithObserver returns an option that registers the given observer for the
// duration of the operation.
//
// Multiple observers can be registered during a single operation.
func WithObserver(o fact.Observer) OperationOption {
	if o == nil {
		panic("observer must not be nil")
	}

	return func(oo *operationOptions) {
		oo.observers = append(oo.observers, o)
	}
}

// EnableAggregates returns an operation option that enables or disables
// aggregate message handlers.
//
// All handler types are enabled by default.
func EnableAggregates(enabled bool) OperationOption {
	return enableHandlerType(handler.AggregateType, enabled)
}

// EnableProcesses returns an operation option that enables or disables process
// message handlers.
//
// All handler types are enabled by default.
func EnableProcesses(enabled bool) OperationOption {
	return enableHandlerType(handler.ProcessType, enabled)
}

// EnableIntegrations returns an operation option that enables or disables
// integration message handlers.
//
// All handler types are enabled by default.
func EnableIntegrations(enabled bool) OperationOption {
	return enableHandlerType(handler.IntegrationType, enabled)
}

// EnableProjections returns an operation option that enables or disables
// projection message handlers.
//
// All handler types are enabled by default.
func EnableProjections(enabled bool) OperationOption {
	return enableHandlerType(handler.ProjectionType, enabled)
}

// enableHandlerType returns an operation option that enables or disables
// handlers of the given type.
func enableHandlerType(t handler.Type, enabled bool) OperationOption {
	t.MustValidate()

	return func(oo *operationOptions) {
		oo.enabledHandlers[t] = enabled
	}
}

// WithCurrentTime returns an operation option that sets the engine's current
// time.
//
// Note that if this option is used with the test runner, it will take
// precedence over the testkit.WithStartTime() option, clobbering any
// adjustments to the test clock made with the Test.AdvanceTimeTo() and
// AdvanceTimeBy() methods.
func WithCurrentTime(t time.Time) OperationOption {
	return func(oo *operationOptions) {
		oo.now = t
	}
}

// operationOptions is a container for the options set via OperationOption values.
type operationOptions struct {
	now             time.Time
	observers       fact.ObserverGroup
	enabledHandlers map[handler.Type]bool
}

// newOperationOptions returns a new operationOptions with the given options.
func newOperationOptions(options []OperationOption) *operationOptions {
	oo := &operationOptions{
		now: time.Now(),
		enabledHandlers: map[handler.Type]bool{
			handler.AggregateType:   true,
			handler.ProcessType:     true,
			handler.IntegrationType: true,
			handler.ProjectionType:  true,
		},
	}

	for _, opt := range options {
		opt(oo)
	}

	return oo
}
