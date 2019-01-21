package config

import (
	"context"
	"reflect"

	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// Config is an interface for all configuration values.
type Config interface {
	// Name returns the name of the configured item.
	// For example, the application or handler name.
	Name() string

	// Accept calls the appropriate method on v for this configuration type.
	Accept(ctx context.Context, v Visitor) error
}

// HandlerConfig is an interface for configuration values that refer to a
// specific message handler.
type HandlerConfig interface {
	Config

	// HandleType returns the type of handler.
	HandlerType() handler.Type

	// HandlerReflectType returns the reflect.Type of the handler.
	HandlerReflectType() reflect.Type

	// CommandTypes returns the types of command messages that are routed to the handler.
	CommandTypes() message.TypeSet

	// EventTypes returns the types of event messages that are routed to the handler.
	EventTypes() message.TypeSet
}
