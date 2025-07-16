package panicx

import (
	"fmt"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/location"
)

// UnexpectedBehavior is a panic value that occurs when a handler exhibits some
// behavior that the engine did not expect.
//
// Often this means it has violated the Dogma specification.
type UnexpectedBehavior struct {
	// Handler is the non-compliant handler.
	Handler config.Handler

	// Interface is the name of the interface containing the method with the
	// unexpected behavior.
	Interface string

	// Method is the name of the method that behaved unexpectedly.
	Method string

	// Implementation is the value that implements the nominated interface.
	Implementation any

	// Message is the message that was being handled at the time, if any.
	Message dogma.Message

	// Description is a human-readable description of the behavior.
	Description string

	// Location is the engine's best attempt at pinpointing the location of the
	// unexpected behavior.
	Location location.Location
}

func (x UnexpectedBehavior) String() string {
	return fmt.Sprintf(
		"the '%s' %s message handler behaved unexpectedly in %T.%s(): %s",
		x.Handler.Identity().Name,
		x.Handler.HandlerType(),
		x.Implementation,
		x.Method,
		x.Description,
	)
}
