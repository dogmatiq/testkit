package engine

import (
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/fact"
)

// OperationOption applies optional settings while dispatching a message or
// performing a tick.
type OperationOption interface {
	applyOperationOption(*operationOptions)
}

type operationOptionFunc func(*operationOptions)

func (f operationOptionFunc) applyOperationOption(opts *operationOptions) {
	f(opts)
}

// WithObserver returns an option that registers the given observer for the
// duration of the operation.
//
// Multiple observers can be registered during a single operation.
func WithObserver(o fact.Observer) OperationOption {
	if o == nil {
		panic("observer must not be nil")
	}

	return operationOptionFunc(func(oo *operationOptions) {
		oo.observers = append(oo.observers, o)
	})
}

// EnableAggregates returns an operation option that enables or disables
// aggregate message handlers.
//
// All handler types are enabled by default.
func EnableAggregates(enabled bool) OperationOption {
	return enableHandlerType(configkit.AggregateHandlerType, enabled)
}

// EnableProcesses returns an operation option that enables or disables process
// message handlers.
//
// All handler types are enabled by default.
func EnableProcesses(enabled bool) OperationOption {
	return enableHandlerType(configkit.ProcessHandlerType, enabled)
}

// EnableIntegrations returns an operation option that enables or disables
// integration message handlers.
//
// All handler types are enabled by default.
func EnableIntegrations(enabled bool) OperationOption {
	return enableHandlerType(configkit.IntegrationHandlerType, enabled)
}

// EnableProjections returns an operation option that enables or disables
// projection message handlers.
//
// All handler types are enabled by default.
func EnableProjections(enabled bool) OperationOption {
	return enableHandlerType(configkit.ProjectionHandlerType, enabled)
}

// enableHandlerType returns an operation option that enables or disables
// handlers of the given type.
func enableHandlerType(t configkit.HandlerType, enabled bool) OperationOption {
	t.MustValidate()

	return operationOptionFunc(func(oo *operationOptions) {
		oo.enabledHandlerTypes[t] = enabled
	})
}

// EnableHandler returns an operation option that enables or disables a specific
// handler.
//
// This option takes precedence over any EnableAggregates(), EnableProcesses(),
// EnableIntegrations() or EnableProjections() options.
func EnableHandler(name string, enabled bool) OperationOption {
	if err := configkit.ValidateIdentityName(name); err != nil {
		panic(err)
	}

	return operationOptionFunc(func(oo *operationOptions) {
		oo.enabledHandlers[name] = enabled
	})
}

// WithCurrentTime returns an operation option that sets the engine's current
// time.
func WithCurrentTime(t time.Time) OperationOption {
	return operationOptionFunc(func(oo *operationOptions) {
		oo.now = t
	})
}

// operationOptions is a container for the options set via OperationOption
// values.
type operationOptions struct {
	now                 time.Time
	observers           fact.ObserverGroup
	enabledHandlerTypes map[configkit.HandlerType]bool
	enabledHandlers     map[string]bool
}

// newOperationOptions returns a new operationOptions with the given options.
func newOperationOptions(options []OperationOption) *operationOptions {
	oo := &operationOptions{
		now: time.Now(),
		enabledHandlerTypes: map[configkit.HandlerType]bool{
			configkit.AggregateHandlerType:   true,
			configkit.ProcessHandlerType:     true,
			configkit.IntegrationHandlerType: true,
			configkit.ProjectionHandlerType:  true,
		},
		enabledHandlers: map[string]bool{},
	}

	for _, opt := range options {
		opt.applyOperationOption(oo)
	}

	return oo
}
