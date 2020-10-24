package engine

import (
	"context"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/controller/aggregate"
	"github.com/dogmatiq/testkit/engine/controller/integration"
	"github.com/dogmatiq/testkit/engine/controller/process"
	"github.com/dogmatiq/testkit/engine/controller/projection"
)

type configurer struct {
	engine *Engine
}

func (c *configurer) VisitRichApplication(ctx context.Context, cfg configkit.RichApplication) error {
	c.engine.roles = cfg.MessageTypes().All()

	return cfg.RichHandlers().AcceptRichVisitor(ctx, c)
}

func (c *configurer) VisitRichAggregate(_ context.Context, cfg configkit.RichAggregate) error {
	mt := cfg.MessageTypes()
	c.registerController(
		aggregate.NewController(
			cfg,
			&c.engine.messageIDs,
			mt.Produced,
		),
		mt.Consumed,
	)

	return nil
}

func (c *configurer) VisitRichProcess(_ context.Context, cfg configkit.RichProcess) error {
	mt := cfg.MessageTypes()
	c.registerController(
		process.NewController(
			cfg,
			&c.engine.messageIDs,
			mt.Produced,
		),
		mt.Consumed,
	)

	return nil
}

func (c *configurer) VisitRichIntegration(_ context.Context, cfg configkit.RichIntegration) error {
	mt := cfg.MessageTypes()
	c.registerController(
		integration.NewController(
			cfg,
			&c.engine.messageIDs,
			mt.Produced,
		),
		mt.Consumed,
	)

	return nil
}

func (c *configurer) VisitRichProjection(_ context.Context, cfg configkit.RichProjection) error {
	mt := cfg.MessageTypes()
	c.registerController(
		projection.NewController(
			cfg.Identity(),
			cfg.Handler(),
		),
		mt.Consumed,
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
