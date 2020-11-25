package fact

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/testkit/envelope"
)

// EventRecordedByIntegration indicates that an integration recorded an event
// while handling a command.
type EventRecordedByIntegration struct {
	Handler       configkit.RichIntegration
	Envelope      *envelope.Envelope
	EventEnvelope *envelope.Envelope
}

// MessageLoggedByIntegration indicates that an integration wrote a log message
// while handling a command.
type MessageLoggedByIntegration struct {
	Handler      configkit.RichIntegration
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
