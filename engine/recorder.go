package engine

import (
	"context"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
)

// EventRecorder adapts an Engine to the dogma.EventRecorder interface.
type EventRecorder struct {
	// Engine is the engine that handles the recorded events.
	Engine *Engine

	// Options is a set of options used when dispatching the message to the
	// engine.
	Options []OperationOption
}

// RecordEvent records the occurrence of an event.
//
// It is not an error to record an event that is not routed to any handlers.
func (r EventRecorder) RecordEvent(ctx context.Context, m dogma.Message) error {
	t := message.TypeOf(m)
	r.Engine.roles[t].MustBe(message.EventRole)

	return r.Engine.Dispatch(ctx, m, r.Options...)
}
