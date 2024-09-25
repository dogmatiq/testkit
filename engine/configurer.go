package engine

import (
	"context"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/engine/internal/aggregate"
	"github.com/dogmatiq/testkit/engine/internal/integration"
	"github.com/dogmatiq/testkit/engine/internal/process"
	"github.com/dogmatiq/testkit/engine/internal/projection"
)

type configurer struct {
	options *engineOptions
	engine  *Engine
}

func (c *configurer) VisitRichApplication(ctx context.Context, cfg configkit.RichApplication) error {
	return cfg.RichHandlers().AcceptRichVisitor(ctx, c)
}

func (c *configurer) VisitRichAggregate(_ context.Context, cfg configkit.RichAggregate) error {
	mt := cfg.MessageTypes()
	c.registerController(
		&aggregate.Controller{
			Config:     cfg,
			MessageIDs: &c.engine.messageIDs,
		},
		mt.Consumed,
	)

	return nil
}

func (c *configurer) VisitRichProcess(_ context.Context, cfg configkit.RichProcess) error {
	mt := cfg.MessageTypes()
	c.registerController(
		&process.Controller{
			Config:     cfg,
			MessageIDs: &c.engine.messageIDs,
		},
		mt.Consumed,
	)

	return nil
}

func (c *configurer) VisitRichIntegration(_ context.Context, cfg configkit.RichIntegration) error {
	mt := cfg.MessageTypes()
	c.registerController(
		&integration.Controller{
			Config:     cfg,
			MessageIDs: &c.engine.messageIDs,
		},
		mt.Consumed,
	)

	return nil
}

func (c *configurer) VisitRichProjection(_ context.Context, cfg configkit.RichProjection) error {
	mt := cfg.MessageTypes()
	c.registerController(
		&projection.Controller{
			Config:                cfg,
			CompactDuringHandling: c.options.compactDuringHandling,
		},
		mt.Consumed,
	)

	return nil
}

func (c *configurer) registerController(
	ctrl controller,
	consumed message.Set[message.Type],
) {
	c.engine.controllers[ctrl.HandlerConfig().Identity().Name] = ctrl

	for t := range consumed.All() {
		c.engine.routes[t] = append(c.engine.routes[t], ctrl)
	}
}
