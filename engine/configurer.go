package engine

import (
	"context"

	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/controller/aggregate"
	"github.com/dogmatiq/dogmatest/engine/controller/integration"
	"github.com/dogmatiq/dogmatest/engine/controller/process"
	"github.com/dogmatiq/dogmatest/engine/controller/projection"
	"github.com/dogmatiq/dogmatest/internal/enginekit/config"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
	"github.com/dogmatiq/dogmatest/render"
)

type configurer struct {
	engine     *Engine
	comparator compare.Comparator
	renderer   render.Renderer
}

func (c *configurer) VisitAppConfig(ctx context.Context, cfg *config.AppConfig) error {
	for _, h := range cfg.Handlers {
		if err := h.Accept(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

func (c *configurer) VisitAggregateConfig(_ context.Context, cfg *config.AggregateConfig) error {
	c.registerController(
		aggregate.NewController(
			cfg.HandlerName,
			cfg.Handler,
		),
		message.CommandRole,
		cfg.CommandTypes,
	)

	return nil
}

func (c *configurer) VisitProcessConfig(_ context.Context, cfg *config.ProcessConfig) error {
	c.registerController(
		process.NewController(
			cfg.HandlerName,
			cfg.Handler,
		),
		message.EventRole,
		cfg.EventTypes,
	)

	return nil
}

func (c *configurer) VisitIntegrationConfig(_ context.Context, cfg *config.IntegrationConfig) error {
	c.registerController(
		integration.NewController(
			cfg.HandlerName,
			cfg.Handler,
		),
		message.CommandRole,
		cfg.CommandTypes,
	)

	return nil
}

func (c *configurer) VisitProjectionConfig(_ context.Context, cfg *config.ProjectionConfig) error {
	c.registerController(
		projection.NewController(
			cfg.HandlerName,
			cfg.Handler,
		),
		message.EventRole,
		cfg.EventTypes,
	)

	return nil
}

func (c *configurer) registerController(
	ctrl controller.Controller,
	r message.Role,
	types map[message.Type]struct{},
) {
	for t := range types {
		c.engine.routes[t] = append(c.engine.routes[t], ctrl)
		c.engine.roles[t] = r
	}
}
