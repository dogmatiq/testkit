package projection

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/engine/internal/panicx"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/fact"
)

// CompactInterval is how frequently projections should be compacted.
//
// This interval respects the current engine time, which may not be the same as
// the "real world" time. See engine.RunTimeScaled().
const CompactInterval = 1 * time.Hour

// Controller is an implementation of engine.Controller for
// dogma.ProjectionMessageHandler implementations.
type Controller struct {
	Config                configkit.RichProjection
	CompactDuringHandling bool

	lastCompact time.Time
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() configkit.RichHandler {
	return c.Config
}

// Tick always performs projection compaction.
func (c *Controller) Tick(
	ctx context.Context,
	obs fact.Observer,
	now time.Time,
) ([]*envelope.Envelope, error) {
	if now.Sub(c.lastCompact) >= CompactInterval {
		c.lastCompact = now

		obs.Notify(fact.ProjectionCompactionBegun{
			Handler: c.Config,
		})

		err := c.Config.Handler().Compact(
			ctx,
			&scope{
				config:   c.Config,
				observer: obs,
				now:      now,
			},
		)

		obs.Notify(fact.ProjectionCompactionCompleted{
			Handler: c.Config,
			Error:   err,
		})

		return nil, err
	}

	return nil, nil
}

// Handle handles a message.
func (c *Controller) Handle(
	ctx context.Context,
	obs fact.Observer,
	now time.Time,
	env *envelope.Envelope,
) ([]*envelope.Envelope, error) {
	if !c.Config.MessageTypes().Consumed.Has(env.Type) {
		panic(fmt.Sprintf("%s does not handle %s messages", c.Config.Identity(), env.Type))
	}

	handler := c.Config.Handler()

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

	compactResult := make(chan error, 1)

	if c.CompactDuringHandling {
		// Ensure that notification of facts occurs in the main goroutine as
		// observers aren't required to be thread-safe.
		obs.Notify(fact.ProjectionCompactionBegun{
			Handler: c.Config,
		})

		// Start a goroutine so that compaction happens in parallel with
		// handling the message. This is intended to ensure the implementation
		// can actually handle such parallelism, which is required by the spec.
		go func() {
			compactResult <- c.Config.Handler().Compact(
				ctx,
				&scope{
					config:   c.Config,
					observer: obs,
					now:      now,
				},
			)
		}()
	} else {
		close(compactResult)
	}

	var ok bool
	panicx.EnrichUnexpectedMessage(
		c.Config,
		"ProjectionMessageHandler",
		"HandleEvent",
		handler,
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

	compactErr := <-compactResult

	if c.CompactDuringHandling {
		obs.Notify(fact.ProjectionCompactionCompleted{
			Handler: c.Config,
			Error:   compactErr,
		})
	}

	if err != nil {
		return nil, err
	}

	// If this call to handle actually applied the event, close the resource as
	// we'll never invoke the handler with this message again.
	if ok {
		if err := handler.CloseResource(ctx, res); err != nil {
			return nil, err
		}
	}

	// Finally we return the compaction error only if there was no other more
	// relevant error.
	return nil, compactErr
}

// Reset does nothing.
func (c *Controller) Reset() {
}
