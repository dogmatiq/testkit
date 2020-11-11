package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
)

// ProjectionCompactionBegun indicates that a projection is about to be
// compacted.
type ProjectionCompactionBegun struct {
	HandlerName string
}

// ProjectionCompactionCompleted indicates that projection compaction has been
// performed, either successfully or unsuccessfully.
type ProjectionCompactionCompleted struct {
	HandlerName string
	Error       error
}

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
