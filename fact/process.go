package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/envelope"
)

// ProcessInstanceLoaded indicates that a process message handler has loaded an
// existing instance in order to handle an event or timeout.
type ProcessInstanceLoaded struct {
	Handler    *config.Process
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// ProcessEventIgnored indicates that a process message handler chose not to
// route an event to any instance.
type ProcessEventIgnored struct {
	Handler  *config.Process
	Envelope *envelope.Envelope
}

// ProcessEventRoutedToEndedInstance indicates that a process message handler
// ignored an event message because it was routed to an instance that no longer
// exists.
type ProcessEventRoutedToEndedInstance struct {
	Handler    *config.Process
	InstanceID string
	Envelope   *envelope.Envelope
}

// ProcessTimeoutRoutedToEndedInstance indicates that a process message handler
// ignored a timeout message because its instance no longer exists.
type ProcessTimeoutRoutedToEndedInstance struct {
	Handler    *config.Process
	InstanceID string
	Envelope   *envelope.Envelope
}

// ProcessInstanceNotFound indicates that a process message handler was unable
// to load an existing instance while handling an event or timeout.
type ProcessInstanceNotFound struct {
	Handler    *config.Process
	InstanceID string
	Envelope   *envelope.Envelope
}

// ProcessInstanceBegun indicates that a process message handler began a process
// instance while handling an event.
type ProcessInstanceBegun struct {
	Handler    *config.Process
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// ProcessInstanceEnded indicates that a process message handler destroyed a
// process instance while handling an event or timeout.
type ProcessInstanceEnded struct {
	Handler    *config.Process
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// ProcessInstanceEndingReverted indicates that a process message handler
// "reverted" ending a process instance by executing a new command or scheduling
// a new timeout.
type ProcessInstanceEndingReverted struct {
	Handler    *config.Process
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// CommandExecutedByProcess indicates that a process executed a command while
// handling an event or timeout.
type CommandExecutedByProcess struct {
	Handler         *config.Process
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	CommandEnvelope *envelope.Envelope
}

// TimeoutScheduledByProcess indicates that a process scheduled a timeout while
// handling an event or timeout.
type TimeoutScheduledByProcess struct {
	Handler         *config.Process
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	TimeoutEnvelope *envelope.Envelope
}

// MessageLoggedByProcess indicates that a process wrote a log message while
// handling an event or timeout.
type MessageLoggedByProcess struct {
	Handler      *config.Process
	InstanceID   string
	Root         dogma.ProcessRoot
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []any
}
