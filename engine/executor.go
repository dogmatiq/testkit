package engine

import (
	"context"
	"slices"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/fact"
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
	var (
		options           = e.Options
		observerSatisfied bool
		observerError     error
		hasObservers      bool
	)

	if len(opts) > 0 {
		options = slices.Clone(options)

		for _, opt := range opts {
			switch o := opt.(type) {
			case dogma.IdempotencyKeyOption:
				options = append(options, WithIdempotencyKey(o.Key()))
			case dogma.EventObserverOption:
				hasObservers = true
				adapted := adaptEventObserver(ctx, o, &observerSatisfied, &observerError)
				options = append(options, WithObserver(adapted))
			}
		}
	}

	if err := e.Engine.Dispatch(ctx, m, options...); err != nil {
		return err
	}

	if !hasObservers {
		return nil
	}

	if observerError != nil {
		return observerError
	}

	if observerSatisfied {
		return nil
	}

	return dogma.ErrEventObserverNotSatisfied
}

// adaptEventObserver adapts a [dogma.EventObserver] to a [fact.Observer].
func adaptEventObserver(
	ctx context.Context,
	opt dogma.EventObserverOption,
	observerSatisfied *bool,
	observerErr *error,
) fact.Observer {
	observedType := opt.EventType().ID()
	observe := opt.Observer()

	return fact.ObserverFunc(func(f fact.Fact) {
		if *observerSatisfied || *observerErr != nil {
			return
		}

		var event dogma.Event

		switch f := f.(type) {
		case fact.EventRecordedByAggregate:
			event = f.EventEnvelope.Message.(dogma.Event)
		case fact.EventRecordedByIntegration:
			event = f.EventEnvelope.Message.(dogma.Event)
		default:
			return
		}

		t, ok := dogma.RegisteredMessageTypeOf(event)
		if !ok || t.ID() != observedType {
			return
		}

		satisfied, err := observe(ctx, event)
		if err != nil {
			*observerErr = err
		} else if satisfied {
			*observerSatisfied = true
		}
	})
}
