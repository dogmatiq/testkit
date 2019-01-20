package fact

import (
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// EventRecordedByIntegration indicates that an integration recorded an event
// while handling a command.
type EventRecordedByIntegration struct {
	HandlerName   string
	Envelope      *envelope.Envelope
	EventEnvelope *envelope.Envelope
}

// MessageLoggedByIntegration indicates that an integration wrote a log message
// while handling a command.
type MessageLoggedByIntegration struct {
	HandlerName  string
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
