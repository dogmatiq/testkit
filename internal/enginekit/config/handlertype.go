package config

import (
	"context"
	"reflect"
)

// handlerType returns the type of the handler that cfg applies to.
// It panics if cfg is an *AppConfig.
func handlerType(cfg Config) reflect.Type {
	var v handlerTypeVisitor

	ctx := context.Background()
	if err := cfg.Accept(ctx, &v); err != nil {
		panic(err)
	}

	return v.Type
}

type handlerTypeVisitor struct {
	Type reflect.Type
}

func (handlerTypeVisitor) VisitAppConfig(_ context.Context, cfg *AppConfig) error {
	panic("not implemented")
}

func (v *handlerTypeVisitor) VisitAggregateConfig(_ context.Context, cfg *AggregateConfig) error {
	v.Type = reflect.TypeOf(cfg.Handler)
	return nil
}

func (v *handlerTypeVisitor) VisitProcessConfig(_ context.Context, cfg *ProcessConfig) error {
	v.Type = reflect.TypeOf(cfg.Handler)
	return nil
}

func (v *handlerTypeVisitor) VisitIntegrationConfig(_ context.Context, cfg *IntegrationConfig) error {
	v.Type = reflect.TypeOf(cfg.Handler)
	return nil
}

func (v *handlerTypeVisitor) VisitProjectionConfig(_ context.Context, cfg *ProjectionConfig) error {
	v.Type = reflect.TypeOf(cfg.Handler)
	return nil
}
