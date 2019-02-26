package engine

import (
	"context"

	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
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
			&c.engine.messageIDs,
			roleMapToSet(cfg.ProducedMessageTypes()),
		),
		cfg.ConsumedMessageTypes(),
	)

	return nil
}

func (c *configurer) VisitProcessConfig(_ context.Context, cfg *config.ProcessConfig) error {
	c.registerController(
		process.NewController(
			cfg.HandlerName,
			cfg.Handler,
			&c.engine.messageIDs,
			roleMapToSet(cfg.ProducedMessageTypes()),
		),
		cfg.ConsumedMessageTypes(),
	)

	return nil
}

func (c *configurer) VisitIntegrationConfig(_ context.Context, cfg *config.IntegrationConfig) error {
	c.registerController(
		integration.NewController(
			cfg.HandlerName,
			cfg.Handler,
			&c.engine.messageIDs,
			roleMapToSet(cfg.ProducedMessageTypes()),
		),
		cfg.ConsumedMessageTypes(),
	)

	return nil
}

func (c *configurer) VisitProjectionConfig(_ context.Context, cfg *config.ProjectionConfig) error {
	c.registerController(
		projection.NewController(
			cfg.HandlerName,
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
	c.engine.controllers[ctrl.Name()] = ctrl

	for t := range types {
		c.engine.routes[t] = append(c.engine.routes[t], ctrl)
	}
}

// roleMapToSet converts a map of message type to role, as used by the enginekit
// config system, into a type set.
func roleMapToSet(m map[message.Type]message.Role) message.TypeSet {
	s := message.TypeSet{}

	for t := range m {
		s.Add(t)
	}

	return s
}
