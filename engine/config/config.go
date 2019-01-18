package config

// Config is an interface for all configuration values.
type Config interface {
	// Name returns the name of the configured item.
	// For example, the application or handler name.
	Name() string

	// Accept calls the appropriate method on for this configuration type.
	Accept(v Visitor)
}

// Visitor is an interface for visitors to configuration values.
type Visitor interface {
	VisitAppConfig(*AppConfig)
	VisitAggregateConfig(*AggregateConfig)
	VisitProcessConfig(*ProcessConfig)
	VisitIntegrationConfig(*IntegrationConfig)
	VisitProjectionConfig(*ProjectionConfig)
}
