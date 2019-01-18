package engine

import "context"

// Controller orchestrates the handling of a message by Dogma message handler.
type Controller interface {
	// Name returns the name of the handler.
	Name() string

	// Handler returns the application's underlying handler.
	Handler() interface{}

	// Handle handles a message.
	Handle(ctx context.Context, env *Envelope) error

	// Reset clears the state of the controller.
	Reset()
}
