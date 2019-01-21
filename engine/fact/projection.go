package fact

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// MessageLoggedByProjection indicates that a projection wrote a log message while
// handling an event.
type MessageLoggedByProjection struct {
	HandlerName  string
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}