package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// AggregateInstanceLoaded indicates that an aggregate message handler has
// loaded an existing instance in order to handle a command.
type AggregateInstanceLoaded struct {
	HandlerName     string
	InstanceID      string
	Root            dogma.AggregateRoot
	CommandEnvelope *envelope.Envelope
}

// AggregateInstanceNotFound indicates that an aggregate message handler was
// unable to load an existing instance while handling a command.
type AggregateInstanceNotFound struct {
	HandlerName     string
	InstanceID      string
	CommandEnvelope *envelope.Envelope
}

// AggregateInstanceCreated indicates that an aggregate message handler created
// an aggregate instance while handling a command.
type AggregateInstanceCreated struct {
	HandlerName     string
	InstanceID      string
	Root            dogma.AggregateRoot
	CommandEnvelope *envelope.Envelope
}

// AggregateInstanceDestroyed indicates that an aggregate message handler
// destroyed an aggregate instance while handling a command.
type AggregateInstanceDestroyed struct {
	HandlerName     string
	InstanceID      string
	Root            dogma.AggregateRoot
	CommandEnvelope *envelope.Envelope
}

// EventRecordedByAggregate indicates that an aggregate recorded an event while
// handling a command.
type EventRecordedByAggregate struct {
	HandlerName     string
	InstanceID      string
	Root            dogma.AggregateRoot
	CommandEnvelope *envelope.Envelope
	EventEnvelope   *envelope.Envelope
}

// AggregateLoggedMessage indicates that an aggregate wrote a log message while
// handling a command.
type AggregateLoggedMessage struct {
	HandlerName     string
	InstanceID      string
	Root            dogma.AggregateRoot
	CommandEnvelope *envelope.Envelope
	LogFormat       string
	LogArguments    []interface{}
}
