package controller

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
)

// UnexpectedMessage is a panic value that provides more context when a handler
// panics with a dogma.UnexpoectedMessage value.
type UnexpectedMessage struct {
	// Handler is the handler that panicked.
	Handler configkit.RichHandler

	// Method is that method of the handler that panicked.
	Method string

	// Message is the message that caused the handler to panic.
	Message dogma.Message
}

// ConvertUnexpectedMessage calls fn() and converts dogma.UnexpectedMessage
// values to an controller.UnexpectedMessage value to provide more context about
// the failure.
func ConvertUnexpectedMessage(
	h configkit.RichHandler,
	method string,
	m dogma.Message,
	fn func(),
) {
	defer func() {
		v := recover()

		if v == dogma.UnexpectedMessage {
			v = UnexpectedMessage{
				Handler: h,
				Method:  method,
				Message: m,
			}
		}

		panic(v)
	}()

	fn()
}
