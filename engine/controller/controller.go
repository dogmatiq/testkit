package controller

import (
	"context"

	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// Controller orchestrates the handling of a message by Dogma message handler.
type Controller interface {
	// Name returns the name of the handler that managed by this controller.
	Name() string

	// Handle handles a message.
	Handle(ctx context.Context, s Scope) ([]*envelope.Envelope, error)

	// Reset clears the state of the controller.
	Reset()
}
