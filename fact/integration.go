package fact

import (
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/envelope"
)

// EventRecordedByIntegration indicates that an integration recorded an event
// while handling a command.
type EventRecordedByIntegration struct {
	Handler       *config.Integration
	Envelope      *envelope.Envelope
	EventEnvelope *envelope.Envelope
}

// MessageLoggedByIntegration indicates that an integration wrote a log message
// while handling a command.
type MessageLoggedByIntegration struct {
	Handler      *config.Integration
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []any
}
