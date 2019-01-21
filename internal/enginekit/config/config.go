package config

import (
	"context"

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
}

// Visitor is an interface for walking application configurations.
type Visitor interface {
	// VisitAppConfig is called by AppConfig.Accept().
	VisitAppConfig(context.Context, *AppConfig) error

	// VisitAggregateConfig is called by AggregateConfig.Accept().
	VisitAggregateConfig(context.Context, *AggregateConfig) error

	// VisitProcessConfig is called by ProcessConfig.Accept().
	VisitProcessConfig(context.Context, *ProcessConfig) error

	// VisitIntegrationConfig is called by IntegrationConfig.Accept().
	VisitIntegrationConfig(context.Context, *IntegrationConfig) error

	// VisitProjectionConfig is called by ProjectionConfig.Accept().
	VisitProjectionConfig(context.Context, *ProjectionConfig) error
}
