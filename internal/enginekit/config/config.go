package config

import (
	"context"
	"reflect"

	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// Config is an interface for all configuration values.
type Config interface {
	// Name returns the name of the configured item.
	// For example, the application or handler name.
	Name() string

	// Accept calls the appropriate method on v for this configuration type.
	Accept(ctx context.Context, v Visitor) error
}

// HandlerConfig is an interface for handler configurations.
type HandlerConfig interface {
	Config

	// HandleType returns the type of handler that the config applies to.
	HandlerType() handler.Type

	// HandlerReflectType returns the reflect.Type of the handler that the config
	// applies to.
	HandlerReflectType() reflect.Type
}
