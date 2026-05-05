package fact

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/testkit/envelope"
)

// ProcessInstanceLoaded indicates that a process message handler has loaded an
// existing instance in order to handle an event or deadline.
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

// ProcessDeadlineRoutedToEndedInstance indicates that a process message handler
// ignored a deadline message because its instance no longer exists.
type ProcessDeadlineRoutedToEndedInstance struct {
	Handler    *config.Process
	InstanceID string
	Envelope   *envelope.Envelope
}

// ProcessInstanceNotFound indicates that a process message handler was unable
// to load an existing instance while handling an event or deadline.
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
// process instance while handling an event or deadline.
type ProcessInstanceEnded struct {
	Handler    *config.Process
	InstanceID string
	Root       dogma.ProcessRoot
	Envelope   *envelope.Envelope
}

// CommandExecutedByProcess indicates that a process executed a command while
// handling an event or deadline.
type CommandExecutedByProcess struct {
	Handler         *config.Process
	InstanceID      string
	Root            dogma.ProcessRoot
	Envelope        *envelope.Envelope
	CommandEnvelope *envelope.Envelope
}

// DeadlineScheduledByProcess indicates that a process scheduled a deadline while
// handling an event or deadline.
type DeadlineScheduledByProcess struct {
	Handler          *config.Process
	InstanceID       string
	Root             dogma.ProcessRoot
	Envelope         *envelope.Envelope
	DeadlineEnvelope *envelope.Envelope
}

// MessageLoggedByProcess indicates that a process wrote a log message while
// handling an event or deadline.
type MessageLoggedByProcess struct {
	Handler      *config.Process
	InstanceID   string
	Root         dogma.ProcessRoot
	Ended        bool
	Envelope     *envelope.Envelope
	LogFormat    string
	LogArguments []any
}
