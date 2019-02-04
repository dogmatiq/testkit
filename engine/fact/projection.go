package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
)

// MessageLoggedByProjection indicates that a projection wrote a log message while
// handling an event.
type MessageLoggedByProjection struct {
	HandlerName  string
	Handler      dogma.ProjectionMessageHandler
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
