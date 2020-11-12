package projection

import (
	"context"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/testkit/engine/controller"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact"
)

// CompactInterval is how frequently projections should be compacted.
//
// This interval respects the current engine time, which may not be the same as
// the "real world" time. See engine.RunTimeScaled().
const CompactInterval = 1 * time.Hour

// Controller is an implementation of engine.Controller for
// dogma.ProjectionMessageHandler implementations.
type Controller struct {
	Config configkit.RichProjection

	lastCompact time.Time
}

// Identity returns the identity of the handler that is managed by this
// controller.
func (c *Controller) Identity() configkit.Identity {
	return c.Config.Identity()
}

// Type returns configkit.ProjectionHandlerType.
func (c *Controller) Type() configkit.HandlerType {
	return configkit.ProjectionHandlerType
}

// Tick always performs projection compaction.
func (c *Controller) Tick(
	ctx context.Context,
	obs fact.Observer,
	now time.Time,
) ([]*envelope.Envelope, error) {
	if now.Sub(c.lastCompact) >= CompactInterval {
		c.lastCompact = now
		return nil, c.compact(ctx, obs)
	}

	return nil, nil
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.Observer,
	_ time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	env.Role.MustBe(message.EventRole)

	handler := c.Config.Handler()

	var t time.Duration
	controller.ConvertUnexpectedMessagePanic(
		c.Config,
		"ProjectionMessageHandler",
		"TimeoutHint",
		env.Message,
		func() {
			t = handler.TimeoutHint(env.Message)
		},
	)

	if t != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, t)
		defer cancel()
	}

	s := &scope{
		config:   c.Config,
		observer: obs,
		event:    env,
	}

	// This implementation attempts to use the full suite of OCC operations
	// including ResourceVersion() and CloseResource() in order to more
	// thoroughly test the projection handler. However, a "real" implementation
	// would likely not need to call ResourceVersion() before every call to
	// HandleEvent().

	// The message ID is used as the resource identifier. When the message is
	// handled, the resource version is updated to a non-empty value, indicating
	// that the message has been processed.
	res := []byte(env.MessageID)
	cur, err := handler.ResourceVersion(ctx, res)
	if err != nil {
		return nil, err
	}

	// If the version is non-empty, this message has already been processed.
	// This would likely never occur as part of regular testing.
	if len(cur) != 0 {
		return nil, nil
	}

	var ok bool
	controller.ConvertUnexpectedMessagePanic(
		c.Config,
		"ProjectionMessageHandler",
		"HandleEvent",
		env.Message,
		func() {
			ok, err = handler.HandleEvent(
				ctx,
				res,
				nil,       // current version
				[]byte{1}, // next version
				s,
				env.Message,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	// If this call to handle actually applied the event, close the resource as
	// we'll never invoke the handler with this message again.
	if ok {
		return nil, handler.CloseResource(ctx, res)
	}

	return nil, nil
}

// Reset does nothing.
func (c *Controller) Reset() {
}

// compact performs projection compaction and records facts about it.
func (c *Controller) compact(ctx context.Context, obs fact.Observer) error {
	obs.Notify(fact.ProjectionCompactionBegun{
		HandlerName: c.Config.Identity().Name,
	})

	err := c.Config.Handler().Compact(
		ctx,
		&scope{
			config:   c.Config,
			observer: obs,
		},
	)

	obs.Notify(fact.ProjectionCompactionCompleted{
		HandlerName: c.Config.Identity().Name,
		Error:       err,
	})

	return err
}
