package fact

import (
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/envelope"
)

// ProjectionCompactionBegun indicates that a projection is about to be
// compacted.
type ProjectionCompactionBegun struct {
	Handler *config.Projection
}

// ProjectionCompactionCompleted indicates that projection compaction has been
// performed, either successfully or unsuccessfully.
type ProjectionCompactionCompleted struct {
	Handler *config.Projection
	Error   error
}

// MessageLoggedByProjection indicates that a projection wrote a log message
// while handling an event or compacting the projection.
//
// Envelope is nil if the message was logged during compaction.
type MessageLoggedByProjection struct {
	Handler      *config.Projection
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []any
}
