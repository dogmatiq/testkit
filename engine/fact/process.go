package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine/envelope"
)

// ProcessInstanceLoaded indicates that a process message handler has loaded an
// existing instance in order to handle an event or timeout.
type ProcessInstanceLoaded struct {
	HandlerName string
	Handler     dogma.ProcessMessageHandler
	InstanceID  string
	Root        dogma.ProcessRoot
	Envelope    *envelope.Envelope
}

// ProcessEventIgnored indicates that a process message handler chose not to
// route an event to any instance.
type ProcessEventIgnored struct {
	HandlerName string
	Handler     dogma.ProcessMessageHandler
	Envelope    *envelope.Envelope
}

// ProcessTimeoutIgnored indicates that a process message handler ignored
// a timeout message because its instance no longer exists.
type ProcessTimeoutIgnored struct {
	HandlerName string
	Handler     dogma.ProcessMessageHandler
	Envelope    *envelope.Envelope
}

// ProcessInstanceNotFound indicates that a process message handler was unable
// to load an existing instance while handling an event or timeout.
type ProcessInstanceNotFound struct {
	HandlerName string
	Handler     dogma.ProcessMessageHandler
	InstanceID  string
	Envelope    *envelope.Envelope
}

// ProcessInstanceBegun indicates that a process message handler began a
// process instance while handling an event.
type ProcessInstanceBegun struct {
	HandlerName string
	Handler     dogma.ProcessMessageHandler
	InstanceID  string
	Root        dogma.ProcessRoot
	Envelope    *envelope.Envelope
}

// ProcessInstanceEnded indicates that a process message handler destroyed
// a process instance while handling an event or timeout.
type ProcessInstanceEnded struct {
	HandlerName string
	Handler     dogma.ProcessMessageHandler
	InstanceID  string
	Root        dogma.ProcessRoot
	Envelope    *envelope.Envelope
}

// CommandExecutedByProcess indicates that a process executed a command while
// handling an event or timeout.
type CommandExecutedByProcess struct {
	HandlerName     string
	Handler         dogma.ProcessMessageHandler
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	CommandEnvelope *envelope.Envelope
}

// TimeoutScheduledByProcess indicates that a process scheduled a timeout while
// handling an event or timeout.
type TimeoutScheduledByProcess struct {
	HandlerName     string
	Handler         dogma.ProcessMessageHandler
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	TimeoutEnvelope *envelope.Envelope
}

// MessageLoggedByProcess indicates that a process wrote a log message while
// handling an event or timeout.
type MessageLoggedByProcess struct {
	HandlerName  string
	Handler      dogma.ProcessMessageHandler
	InstanceID   string
	Root         dogma.ProcessRoot
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
