package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// AggregateInstanceLoaded indicates that an aggregate message handler has
// loaded an existing instance in order to handle a command.
type AggregateInstanceLoaded struct {
	HandlerName string
	Handler     dogma.AggregateMessageHandler
	InstanceID  string
	Root        dogma.AggregateRoot
	Envelope    *envelope.Envelope
}

// AggregateInstanceNotFound indicates that an aggregate message handler was
// unable to load an existing instance while handling a command.
type AggregateInstanceNotFound struct {
	HandlerName string
	Handler     dogma.AggregateMessageHandler
	InstanceID  string
	Envelope    *envelope.Envelope
}

// AggregateInstanceCreated indicates that an aggregate message handler created
// an aggregate instance while handling a command.
type AggregateInstanceCreated struct {
	HandlerName string
	Handler     dogma.AggregateMessageHandler
	InstanceID  string
	Root        dogma.AggregateRoot
	Envelope    *envelope.Envelope
}

// AggregateInstanceDestroyed indicates that an aggregate message handler
// destroyed an aggregate instance while handling a command.
type AggregateInstanceDestroyed struct {
	HandlerName string
	Handler     dogma.AggregateMessageHandler
	InstanceID  string
	Root        dogma.AggregateRoot
	Envelope    *envelope.Envelope
}

// EventRecordedByAggregate indicates that an aggregate recorded an event while
// handling a command.
type EventRecordedByAggregate struct {
	HandlerName   string
	Handler       dogma.AggregateMessageHandler
	InstanceID    string
	Root          dogma.AggregateRoot
	Envelope      *envelope.Envelope
	EventEnvelope *envelope.Envelope
}

// MessageLoggedByAggregate indicates that an aggregate wrote a log message
// while handling a command.
type MessageLoggedByAggregate struct {
	HandlerName  string
	Handler      dogma.AggregateMessageHandler
	InstanceID   string
	Root         dogma.AggregateRoot
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
