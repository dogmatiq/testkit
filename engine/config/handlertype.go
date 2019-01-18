package config

import "reflect"

// handlerType returns the type of the handler that cfg applies to.
// It panics if cfg is an *AppConfig.
func handlerType(cfg Config) reflect.Type {
	var v handlerTypeVisitor
	cfg.Accept(&v)
	return v.Type
}

type handlerTypeVisitor struct {
	Type reflect.Type
}

func (handlerTypeVisitor) VisitAppConfig(cfg *AppConfig) {
	panic("not implemented")
}

func (v *handlerTypeVisitor) VisitAggregateConfig(cfg *AggregateConfig) {
	v.Type = reflect.TypeOf(cfg.Handler)
}

func (v *handlerTypeVisitor) VisitProcessConfig(cfg *ProcessConfig) {
	v.Type = reflect.TypeOf(cfg.Handler)
}

func (v *handlerTypeVisitor) VisitIntegrationConfig(cfg *IntegrationConfig) {
	v.Type = reflect.TypeOf(cfg.Handler)
}

func (v *handlerTypeVisitor) VisitProjectionConfig(cfg *ProjectionConfig) {
	v.Type = reflect.TypeOf(cfg.Handler)
}
