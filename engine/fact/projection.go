package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
)

// MessageLoggedByProjection indicates that a projection wrote a log message
// while handling an event or compacting the projection.
//
// Envelope is nil if the message was logged during compaction.
type MessageLoggedByProjection struct {
	HandlerName  string
	Handler      dogma.ProjectionMessageHandler
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
