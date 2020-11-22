package panicx

import (
	"fmt"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
)

// UnexpectedMessage is a panic value that provides more context when a handler
// panics with a dogma.UnexpoectedMessage value.
type UnexpectedMessage struct {
	// Handler is the handler that panicked.
	Handler configkit.RichHandler

	// Interface is the name of the interface containing the method that the
	// controller called resulting in the panic.
	Interface string

	// Method is the name of the method that the controller called resulting in
	// the panic.
	Method string

	// Message is the message that caused the handler to panic.
	Message dogma.Message

	// PanicLocation is the location of the function that panicked, if known.
	PanicLocation Location
}

func (x UnexpectedMessage) String() string {
	return fmt.Sprintf(
		"the '%s' %s message handler did not expect %s() to be called with a message of type %T",
		x.Handler.Identity().Name,
		x.Handler.HandlerType(),
		x.Method,
		x.Message,
	)
}

// EnrichUnexpectedMessage calls fn() and converts dogma.UnexpectedMessage
// values to an controller.UnexpectedMessage value to provide more context about
// the failure.
func EnrichUnexpectedMessage(
	h configkit.RichHandler,
	iface, method string,
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
				Handler:       h,
				Interface:     iface,
				Method:        method,
				Message:       m,
				PanicLocation: LocationOfPanic(),
			}
		}

		panic(v)
	}()

	fn()
}
