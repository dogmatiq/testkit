package engine

import (
	"context"

	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// Controller orchestrates the handling of a message by Dogma message handler.
type Controller interface {
	// Handle handles a message.
	Handle(ctx context.Context, env *envelope.Envelope) error

	// Reset clears the state of the controller.
	Reset()
}
