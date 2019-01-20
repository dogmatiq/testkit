package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// ProcessInstanceLoaded indicates that an process message handler has loaded an
// existing instance in order to handle an event or timeout.
type ProcessInstanceLoaded struct {
	HandlerName string
	InstanceID  string
	Root        dogma.ProcessRoot
	Envelope    *envelope.Envelope
}

// ProcessEventIgnored indicates that an process message handler chose not to
// route an event to any instance.
type ProcessEventIgnored struct {
	HandlerName string
	Envelope    *envelope.Envelope
}

// ProcessInstanceNotFound indicates that an process message handler was unable
// to load an existing instance while handling an event or timeout.
type ProcessInstanceNotFound struct {
	HandlerName string
	InstanceID  string
	Envelope    *envelope.Envelope
}

// ProcessInstanceBegun indicates that an process message handler began a
// process instance while handling an event.
type ProcessInstanceBegun struct {
	HandlerName string
	InstanceID  string
	Root        dogma.ProcessRoot
	Envelope    *envelope.Envelope
}

// ProcessInstanceEnded indicates that an process message handler destroyed
// a process instance while handling an event or timeout.
type ProcessInstanceEnded struct {
	HandlerName string
	InstanceID  string
	Root        dogma.ProcessRoot
	Envelope    *envelope.Envelope
}

// CommandExecutedByProcess indicates that a process executed a command while
// handling an event or timeout.
type CommandExecutedByProcess struct {
	HandlerName     string
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	CommandEnvelope *envelope.Envelope
}

// TimeoutScheduledByProcess indicates that a process scheduled a timeout while
// handling an event or timeout.
type TimeoutScheduledByProcess struct {
	HandlerName     string
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	CommandEnvelope *envelope.Envelope
}

// MessageLoggedByProcess indicates that a process wrote a log message while
// handling an event or timeout.
type MessageLoggedByProcess struct {
	HandlerName  string
	InstanceID   string
	Root         dogma.ProcessRoot
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
