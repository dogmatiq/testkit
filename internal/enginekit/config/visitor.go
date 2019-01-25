package config

import (
	"context"
)

// Visitor is an interface for walking application configurations.
type Visitor interface {
	// VisitApplicationConfig is called by ApplicationConfig.Accept().
	VisitApplicationConfig(context.Context, *ApplicationConfig) error

	// VisitAggregateConfig is called by AggregateConfig.Accept().
	VisitAggregateConfig(context.Context, *AggregateConfig) error

	// VisitProcessConfig is called by ProcessConfig.Accept().
	VisitProcessConfig(context.Context, *ProcessConfig) error

	// VisitIntegrationConfig is called by IntegrationConfig.Accept().
	VisitIntegrationConfig(context.Context, *IntegrationConfig) error

	// VisitProjectionConfig is called by ProjectionConfig.Accept().
	VisitProjectionConfig(context.Context, *ProjectionConfig) error
}

// FuncVisitor is an implementation of Visitor that dispatches to regular
// functions.
type FuncVisitor struct {
	// ApplicationConfig, if non-nil, is called by VisitApplicationConfig().
	ApplicationConfig func(context.Context, *ApplicationConfig) error

	// AggregateConfig, if non-nil, is called by VisitAggregateConfig().
	AggregateConfig func(context.Context, *AggregateConfig) error

	// ProcessConfig, if non-nil, is called by VisitProcessConfig().
	ProcessConfig func(context.Context, *ProcessConfig) error

	// IntegrationConfig, if non-nil, is called by VisitIntegrationConfig().
	IntegrationConfig func(context.Context, *IntegrationConfig) error

	// ProjectionConfig, if non-nil, is called by VisitProjectionConfig().
	ProjectionConfig func(context.Context, *ProjectionConfig) error
}

// VisitApplicationConfig calls v.ApplicationConfig if it is non-nil.
func (v *FuncVisitor) VisitApplicationConfig(ctx context.Context, cfg *ApplicationConfig) error {
	if v.ApplicationConfig != nil {
		return v.ApplicationConfig(ctx, cfg)
	}

	return nil
}

// VisitAggregateConfig calls v.AggregateConfig if it is non-nil.
func (v *FuncVisitor) VisitAggregateConfig(ctx context.Context, cfg *AggregateConfig) error {
	if v.AggregateConfig != nil {
		return v.AggregateConfig(ctx, cfg)
	}

	return nil
}

// VisitProcessConfig calls v.ProcessConfig if it is non-nil.
func (v *FuncVisitor) VisitProcessConfig(ctx context.Context, cfg *ProcessConfig) error {
	if v.ProcessConfig != nil {
		return v.ProcessConfig(ctx, cfg)
	}

	return nil
}

// VisitIntegrationConfig calls v.IntegrationConfig if it is non-nil.
func (v *FuncVisitor) VisitIntegrationConfig(ctx context.Context, cfg *IntegrationConfig) error {
	if v.IntegrationConfig != nil {
		return v.IntegrationConfig(ctx, cfg)
	}

	return nil
}

// VisitProjectionConfig calls v.ProjectionConfig if it is non-nil.
func (v *FuncVisitor) VisitProjectionConfig(ctx context.Context, cfg *ProjectionConfig) error {
	if v.ProjectionConfig != nil {
		return v.ProjectionConfig(ctx, cfg)
	}

	return nil
}
