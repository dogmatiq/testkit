package controller

import (
	"context"

	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
)

// Controller orchestrates the handling of a message by Dogma message handler.
type Controller interface {
	// Name returns the name of the handler that is managed by this controller.
	Name() string

	// Type returns the name of the handler that is managed by this controller.
	Type() handler.Type

	// Handle handles a message.
	Handle(
		ctx context.Context,
		obs fact.Observer,
		env *envelope.Envelope,
	) ([]*envelope.Envelope, error)

	// Reset clears the state of the controller.
	Reset()
}
