package config

import "context"

// Config is an interface for all configuration values.
type Config interface {
	// Name returns the name of the configured item.
	// For example, the application or handler name.
	Name() string

	// Accept calls the appropriate method on for this configuration type.
	Accept(ctx context.Context, v Visitor) error
}

// Visitor is an interface for visitors to configuration values.
type Visitor interface {
	VisitAppConfig(context.Context, *AppConfig) error
	VisitAggregateConfig(context.Context, *AggregateConfig) error
	VisitProcessConfig(context.Context, *ProcessConfig) error
	VisitIntegrationConfig(context.Context, *IntegrationConfig) error
	VisitProjectionConfig(context.Context, *ProjectionConfig) error
}
