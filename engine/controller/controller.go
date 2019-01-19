package controller

import (
	"context"
)

// Controller orchestrates the handling of a message by Dogma message handler.
type Controller interface {
	// Handle handles a message.
	Handle(ctx context.Context, s Scope) error

	// Reset clears the state of the controller.
	Reset()
}
