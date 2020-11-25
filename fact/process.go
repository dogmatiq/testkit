package fact

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
)

// ProcessInstanceLoaded indicates that a process message handler has loaded an
// existing instance in order to handle an event or timeout.
type ProcessInstanceLoaded struct {
	Handler    configkit.RichProcess
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// ProcessEventIgnored indicates that a process message handler chose not to
// route an event to any instance.
type ProcessEventIgnored struct {
	Handler  configkit.RichProcess
	Envelope *envelope.Envelope
}

// ProcessTimeoutIgnored indicates that a process message handler ignored a
// timeout message because its instance no longer exists.
type ProcessTimeoutIgnored struct {
	Handler    configkit.RichProcess
	InstanceID string
	Envelope   *envelope.Envelope
}

// ProcessInstanceNotFound indicates that a process message handler was unable
// to load an existing instance while handling an event or timeout.
type ProcessInstanceNotFound struct {
	Handler    configkit.RichProcess
	InstanceID string
	Envelope   *envelope.Envelope
}

// ProcessInstanceBegun indicates that a process message handler began a process
// instance while handling an event.
type ProcessInstanceBegun struct {
	Handler    configkit.RichProcess
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// ProcessInstanceEnded indicates that a process message handler destroyed a
// process instance while handling an event or timeout.
type ProcessInstanceEnded struct {
	Handler    configkit.RichProcess
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// CommandExecutedByProcess indicates that a process executed a command while
// handling an event or timeout.
type CommandExecutedByProcess struct {
	Handler         configkit.RichProcess
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	CommandEnvelope *envelope.Envelope
}

// TimeoutScheduledByProcess indicates that a process scheduled a timeout while
// handling an event or timeout.
type TimeoutScheduledByProcess struct {
	Handler         configkit.RichProcess
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	TimeoutEnvelope *envelope.Envelope
}

// MessageLoggedByProcess indicates that a process wrote a log message while
// handling an event or timeout.
type MessageLoggedByProcess struct {
	Handler      configkit.RichProcess
	InstanceID   string
	Root         dogma.ProcessRoot
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []interface{}
}
