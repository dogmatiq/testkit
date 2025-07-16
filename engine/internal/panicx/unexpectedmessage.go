package panicx

import (
	"fmt"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/location"
)

// UnexpectedMessage is a panic value that provides more context when a handler
// panics with a dogma.UnexpectedMessage value.
type UnexpectedMessage struct {
	// Handler is the handler that panicked.
	Handler config.Handler

	// Interface is the name of the interface containing the method that the
	// controller called resulting in the panic.
	Interface string

	// Implementation is the value that implements the nominated interface.
	Implementation any

	// Method is the name of the method that the controller called resulting in
	// the panic.
	Method string

	// Message is the message that caused the handler to panic.
	Message dogma.Message

	// PanicLocation is the location of the function that panicked, if known.
	PanicLocation location.Location
}

func (x UnexpectedMessage) String() string {
	return fmt.Sprintf(
		"the '%s' %s message handler did not expect %T.%s() to be called with a message of type %s",
		x.Handler.Identity().Name,
		x.Handler.HandlerType(),
		x.Implementation,
		x.Method,
		message.TypeOf(x.Message),
	)
}

// EnrichUnexpectedMessage calls fn() and converts dogma.UnexpectedMessage
// values to an panicx.UnexpectedMessage value to provide more context about the
// failure.
func EnrichUnexpectedMessage(
	h config.Handler,
	iface string, method string,
	impl any,
	m dogma.Message,
	fn func(),
) {
	defer func() {
		v := recover()

		if v == nil {
			return
		}

		if v == dogma.UnexpectedMessage {
			v = UnexpectedMessage{
				Handler:        h,
				Interface:      iface,
				Method:         method,
				Implementation: impl,
				Message:        m,
				PanicLocation:  location.OfPanic(),
			}
		}

		panic(v)
	}()

	fn()
}
