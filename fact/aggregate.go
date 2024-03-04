package fact

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/envelope"
)

// AggregateInstanceLoaded indicates that an aggregate message handler has
// loaded an existing instance in order to handle a command.
type AggregateInstanceLoaded struct {
	Handler    configkit.RichAggregate
	InstanceID string
	Root       dogma.AggregateRoot
	Envelope   *envelope.Envelope
}

// AggregateInstanceNotFound indicates that an aggregate message handler was
// unable to load an existing instance while handling a command.
type AggregateInstanceNotFound struct {
	Handler    configkit.RichAggregate
	InstanceID string
	Envelope   *envelope.Envelope
}

// AggregateInstanceCreated indicates that an aggregate message handler created
// an aggregate instance while handling a command.
type AggregateInstanceCreated struct {
	Handler    configkit.RichAggregate
	InstanceID string
	Root       dogma.AggregateRoot
	Envelope   *envelope.Envelope
}

// AggregateInstanceDestroyed indicates that an aggregate message handler
// destroyed an aggregate instance while handling a command.
type AggregateInstanceDestroyed struct {
	Handler    configkit.RichAggregate
	InstanceID string
	Root       dogma.AggregateRoot
	Envelope   *envelope.Envelope
}

// AggregateInstanceDestructionReverted indicates that an aggregate message
// handler "reverted" destruction of an aggregate instance by recording a new
// event.
type AggregateInstanceDestructionReverted struct {
	Handler    configkit.RichAggregate
	InstanceID string
	Root       dogma.AggregateRoot
	Envelope   *envelope.Envelope
}

// EventRecordedByAggregate indicates that an aggregate recorded an event while
// handling a command.
type EventRecordedByAggregate struct {
	Handler       configkit.RichAggregate
	InstanceID    string
	Root          dogma.AggregateRoot
	Envelope      *envelope.Envelope
	EventEnvelope *envelope.Envelope
}

// MessageLoggedByAggregate indicates that an aggregate wrote a log message
// while handling a command.
type MessageLoggedByAggregate struct {
	Handler      configkit.RichAggregate
	InstanceID   string
	Root         dogma.AggregateRoot
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []any
}
