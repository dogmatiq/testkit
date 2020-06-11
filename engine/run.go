package engine

import (
	"context"
	"time"

	"github.com/dogmatiq/linger"
)

// DefaultTickInterval is the default interval at which Run() and
// RunTimeScaled() will perform an engine tick.
const DefaultTickInterval = 250 * time.Millisecond

// Run repeatedly calls e.Tick() until ctx is canceled or an error occurs.
//
// d is the duration between ticks. If it is 0, DefaultTickInterval is used.
func Run(
	ctx context.Context,
	e *Engine,
	d time.Duration,
	opts ...OperationOption,
) error {
	return RunTimeScaled(
		ctx,
		e,
		d,
		1,
		time.Time{},
		opts...,
	)
}

// RunTimeScaled repeatedly calls e.Tick() until ctx is canceled or an error
// occurs.
//
// d is the duration between ticks. If it is 0, DefaultTickInterval is used.
//
// Each tick is performed using a WithCurrentTime() option that scales time by a
// factor of f. For example, if f is 2.0, the engine will see time progress by 2
// for every 1 second of real time.
//
// t is the "epoch time", used as current time for the first tick. If t.IsZero()
// is true, the current time is used.
func RunTimeScaled(
	ctx context.Context,
	e *Engine,
	d time.Duration,
	f float64,
	t time.Time,
	opts ...OperationOption,
) error {
	if t.IsZero() {
		t = time.Now()
	}

	// Add a slot at the start of the options for the WithCurrentTime() option.
	opts = append(
		[]OperationOption{nil},
		opts...,
	)

	for {
		elapsed := linger.Multiply(time.Since(t), f)
		opts[0] = WithCurrentTime(t.Add(elapsed))

		if err := e.Tick(ctx, opts...); err != nil {
			return err
		}

		if err := linger.Sleep(ctx, d, DefaultTickInterval); err != nil {
			return err
		}
	}
}
