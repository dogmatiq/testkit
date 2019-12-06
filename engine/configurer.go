package engine

import (
	"context"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/controller/aggregate"
	"github.com/dogmatiq/testkit/engine/controller/integration"
	"github.com/dogmatiq/testkit/engine/controller/process"
	"github.com/dogmatiq/testkit/engine/controller/projection"
)

type configurer struct {
	engine *Engine
}

func (c *configurer) VisitApplicationConfig(ctx context.Context, cfg *config.ApplicationConfig) error {
	c.engine.roles = cfg.Roles

	for _, h := range cfg.HandlersByName {
		if err := h.Accept(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

func (c *configurer) VisitAggregateConfig(_ context.Context, cfg *config.AggregateConfig) error {
	c.registerController(
		aggregate.NewController(
			cfg.HandlerIdentity,
			cfg.Handler,
			&c.engine.messageIDs,
			cfg.ProducedMessageTypes(),
		),
		cfg.ConsumedMessageTypes(),
	)

	return nil
}

func (c *configurer) VisitProcessConfig(_ context.Context, cfg *config.ProcessConfig) error {
	c.registerController(
		process.NewController(
			cfg.HandlerIdentity,
			cfg.Handler,
			&c.engine.messageIDs,
			cfg.ProducedMessageTypes(),
		),
		cfg.ConsumedMessageTypes(),
	)

	return nil
}

func (c *configurer) VisitIntegrationConfig(_ context.Context, cfg *config.IntegrationConfig) error {
	c.registerController(
		integration.NewController(
			cfg.HandlerIdentity,
			cfg.Handler,
			&c.engine.messageIDs,
			cfg.ProducedMessageTypes(),
		),
		cfg.ConsumedMessageTypes(),
	)

	return nil
}

func (c *configurer) VisitProjectionConfig(_ context.Context, cfg *config.ProjectionConfig) error {
	c.registerController(
		projection.NewController(
			cfg.HandlerIdentity,
			cfg.Handler,
		),
		cfg.ConsumedMessageTypes(),
	)

	return nil
}

func (c *configurer) registerController(
	ctrl controller.Controller,
	types map[message.Type]message.Role,
) {
	c.engine.controllers[ctrl.Identity().Name] = ctrl

	for t := range types {
		c.engine.routes[t] = append(c.engine.routes[t], ctrl)
	}
}
