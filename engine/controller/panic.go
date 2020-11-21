package controller

import (
	"fmt"
	"runtime"
	"strings"

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

	// PanicFunc is the name of the function that panicked, if known.
	PanicFunc string

	// PanicFile is the name of the file where the panic originated, if known.
	PanicFile string

	// PanicLine is the line number within the file where the panic originated,
	// if known.
	PanicLine int
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

// ConvertUnexpectedMessagePanic calls fn() and converts dogma.UnexpectedMessage
// values to an controller.UnexpectedMessage value to provide more context about
// the failure.
func ConvertUnexpectedMessagePanic(
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
			name, file, line := findPanicSite()

			v = UnexpectedMessage{
				Handler:   h,
				Interface: iface,
				Method:    method,
				Message:   m,
				PanicFunc: name,
				PanicFile: file,
				PanicLine: line,
			}
		}

		panic(v)
	}()

	fn()
}

func findPanicSite() (string, string, int) {
	var (
		name, file string
		line       int
		pc         [16]uintptr
	)

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)

		if fn != nil {
			name = fn.Name()
			if !strings.HasPrefix(name, "runtime.") {
				file, line = fn.FileLine(pc)
				break
			}
		}
	}

	return name, file, line
}
