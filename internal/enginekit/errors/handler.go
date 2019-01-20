package errors

import "fmt"

// EmptyInstanceID indicates that an aggregate or process message handler has
// attempted to route a message to an instance with an empty ID.
type EmptyInstanceID struct {
	Handler string
}

func (e EmptyInstanceID) Error() string {
	return fmt.Sprintf(
		"the '%s' handler attempted to route a message to an empty instance ID",
		e.Handler,
	)
}

// NilRoot indicates that an aggregate or process message handler has
// returned a nil "root" value from its New() method.
type NilRoot struct {
	Handler string
}

func (e NilRoot) Error() string {
	return fmt.Sprintf(
		"the '%s' handler returned a nil value from New()",
		e.Handler,
	)
}
