package engine

import (
	"context"
	"slices"

	"github.com/dogmatiq/dogma"
)

// CommandExecutor adapts an Engine to the dogma.CommandExecutor interface.
type CommandExecutor struct {
	// Engine is the engine that handles the recorded events.
	Engine *Engine

	// Options is a set of options used when dispatching the message to the
	// engine.
	Options []OperationOption
}

// ExecuteCommand enqueues a command for execution.
//
// It panics if the command is not routed to any handlers.
func (e CommandExecutor) ExecuteCommand(
	ctx context.Context,
	m dogma.Command,
	opts ...dogma.ExecuteCommandOption,
) error {
	options := e.Options
	if len(opts) > 0 {
		options = slices.Clone(options)

		for _, opt := range opts {
			switch o := opt.(type) {
			case dogma.IdempotencyKeyOption:
				options = append(options, WithIdempotencyKey(o.Key()))
			}
		}
	}

	return e.Engine.Dispatch(ctx, m, options...)
}
