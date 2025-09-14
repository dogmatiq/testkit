package projection

import (
	"context"
	"fmt"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
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
	Config                *config.Projection
	CompactDuringHandling bool

	lastCompact time.Time
	checkpoints map[string]uint64
}

// HandlerConfig returns the config of the handler that is managed by this
// controller.
func (c *Controller) HandlerConfig() config.Handler {
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

		err := c.Config.Source.Get().Compact(
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
	mt := message.TypeOf(env.Message)

	if !c.Config.RouteSet().DirectionOf(mt).Has(config.InboundDirection) {
		panic(fmt.Sprintf("%s does not handle %s messages", c.Config.Identity(), mt))
	}

	handler := c.Config.Source.Get()

	// This implementation attempts to use the full suite of OCC operations
	// including CheckpointOffset() to more thoroughly test the projection
	// handler. However, a "real" implementation would likely not need to call
	// CheckpointOffset() before every call to HandleEvent().
	cp, err := handler.CheckpointOffset(ctx, env.EventStreamID)
	if err != nil {
		return nil, err
	}

	// If the checkpoint offset is greater than this event's offset, this
	// message has already been processed.
	if cp > env.EventStreamOffset {
		return nil, nil
	}

	s := &scope{
		config:     c.Config,
		observer:   obs,
		event:      env,
		checkpoint: cp,
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
			compactResult <- c.Config.Source.Get().Compact(
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

	panicx.EnrichUnexpectedMessage(
		c.Config,
		"ProjectionMessageHandler",
		"HandleEvent",
		handler,
		env.Message,
		func() {
			cp, err = handler.HandleEvent(
				ctx,
				s,
				env.Message.(dogma.Event),
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

	if expect := env.EventStreamOffset + 1; cp != expect {
		return nil, fmt.Errorf(
			"optimistic concurrency conflict when handling event at offset %d of stream %s: expected checkpoint offset of %d, handler returned %d",
			env.EventStreamOffset,
			env.EventStreamID,
			expect,
			cp,
		)
	}

	// Finally we return the compaction error only if there was no other more
	// relevant error.
	return nil, compactErr
}

// Reset does nothing.
func (c *Controller) Reset() {
}
