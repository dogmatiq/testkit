package fact

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"uuid"

	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/testkit/envelope"
	"github.com/dogmatiq/testkit/internal/inflect"
)

// Logger is an observer that describes facts in human-readable text. It calls
// a log function once for each fact that it describes, and does nothing for
// the rest.
//
// It describes the following facts only if Trace is true:
//
//   - the engine beginning to dispatch a message or to tick
//   - the engine loading an aggregate or process instance
//   - the engine not finding an aggregate or process instance
//   - an aggregate instance being created
//   - a process instance beginning
//   - the engine skipping a handler
//   - the engine ignoring a message for a process instance that has ended
//
// The engine delivers one message at a time to all of the handlers that route
// it. The logger writes a heading for each message that the engine delivers,
// followed by an indented line for each fact about a handler handling that
// message. For example:
//
//	3d2b4e33  OpenAccount command
//	          aggregate    recorded   account A001 ● *events.AccountOpened ● opened account A001
//	a0864daf  AccountOpened event
//	          process      ignored    open-account-for-new-customer ● not routed to any instance
//
// A heading is the first 8 hexadecimal digits of the message's ID, followed by
// the name of the message's type, without its package, and the kind of the
// message.
//
// An indented line has three columns:
//
//   - the type of the handler, such as "aggregate" or "process"
//   - a word or phrase that says what happened, such as "recorded" or "failed"
//   - the details, starting with the handler's name and, if there is one, the
//     ID of the instance. A "●" separates the parts of the details.
//
// A line about the engine's own work that is not a heading, such as the engine
// ignoring a duplicate command, is not indented. It starts with the ID of the
// message that it is about, if there is one, followed by a word that says what
// happened and the details.
//
// If Trace is true, the ID column of each heading and each line about the
// engine's own work also shows how the message relates to the others. If the
// message is not the one that started the chain of messages, the column starts
// with the ID of the message that started the chain (the correlation ID),
// followed by the ID of the message that caused this one (the causation ID)
// and an arrow. For example:
//
//	                     3d2b4e33  OpenAccount command
//	3d2b4e33  3d2b4e33 → a0864daf  AccountOpened event
type Logger struct {
	// Log is the function that the logger calls with the text of each line.
	Log func(string)

	// Trace makes the logger also describe the facts that the Logger
	// documentation lists as described only if Trace is true, such as the
	// engine loading an instance.
	Trace bool
}

// NewLogger returns a new observer that logs human-readable descriptions of
// facts to the given log function.
func NewLogger(log func(string)) *Logger {
	return &Logger{
		Log: log,
	}
}

// Notify logs a line that describes f, unless f is one of the facts that the
// logger describes only if Trace is true and Trace is false.
func (l *Logger) Notify(f Fact) {
	if !l.Trace && isTraceOnly(f) {
		return
	}

	switch x := f.(type) {
	case DispatchCycleBegun:
		l.dispatchCycleBegun(x)
	case DispatchCycleSkipped:
		l.dispatchCycleSkipped(x)
	case DispatchBegun:
		l.dispatchBegun(x)
	case CommandDeduplicated:
		l.commandDeduplicated(x)
	case HandlingCompleted:
		l.handlingCompleted(x)
	case HandlingSkipped:
		l.handlingSkipped(x)
	case TickCycleBegun:
		l.tickCycleBegun(x)
	case TickCompleted:
		l.tickCompleted(x)
	case AggregateInstanceLoaded:
		l.aggregateInstanceLoaded(x)
	case AggregateInstanceNotFound:
		l.aggregateInstanceNotFound(x)
	case AggregateInstanceCreated:
		l.aggregateInstanceCreated(x)
	case EventRecordedByAggregate:
		l.eventRecordedByAggregate(x)
	case MessageLoggedByAggregate:
		l.messageLoggedByAggregate(x)
	case ProcessInstanceLoaded:
		l.processInstanceLoaded(x)
	case ProcessEventIgnored:
		l.processEventIgnored(x)
	case ProcessEventRoutedToEndedInstance:
		l.processEventRoutedToEndedInstance(x)
	case ProcessDeadlineRoutedToEndedInstance:
		l.processDeadlineRoutedToEndedInstance(x)
	case ProcessInstanceNotFound:
		l.processInstanceNotFound(x)
	case ProcessInstanceBegun:
		l.processInstanceBegun(x)
	case ProcessInstanceEnded:
		l.processInstanceEnded(x)
	case CommandExecutedByProcess:
		l.commandExecutedByProcess(x)
	case DeadlineScheduledByProcess:
		l.deadlineScheduledByProcess(x)
	case MessageLoggedByProcess:
		l.messageLoggedByProcess(x)
	case EventRecordedByIntegration:
		l.eventRecordedByIntegration(x)
	case MessageLoggedByIntegration:
		l.messageLoggedByIntegration(x)
	case ProjectionCompactionCompleted:
		l.projectionCompactionCompleted(x)
	case MessageLoggedByProjection:
		l.messageLoggedByProjection(x)
	}
}

// isTraceOnly returns true if the logger describes f only when Trace is true.
func isTraceOnly(f Fact) bool {
	switch f.(type) {
	case DispatchCycleBegun,
		TickCycleBegun,
		AggregateInstanceLoaded,
		AggregateInstanceNotFound,
		AggregateInstanceCreated,
		ProcessInstanceLoaded,
		ProcessInstanceNotFound,
		ProcessInstanceBegun,
		HandlingSkipped,
		ProcessEventRoutedToEndedInstance,
		ProcessDeadlineRoutedToEndedInstance:
		return true
	default:
		return false
	}
}

// dispatchCycleBegun logs a line that describes f.
func (l *Logger) dispatchCycleBegun(f DispatchCycleBegun) {
	l.logEngine(
		f.Envelope,
		"dispatched",
		formatEngineTime(f.EngineTime),
		formatEnabledHandlers(f.EnabledHandlerTypes, f.EnabledHandlers),
	)
}

// dispatchCycleSkipped logs a line that describes f.
func (l *Logger) dispatchCycleSkipped(f DispatchCycleSkipped) {
	mt := message.TypeOf(f.Envelope.Message)

	var reason string

	switch f.Reason {
	case UnroutedMessageType:
		reason = "no handlers within the application have routes for it"
	case OutboundOnlyMessageType:
		reason = inflect.Sprint(mt.Kind(), "no handlers within the application have inbound routes for it, it is only ever <produced>")
	}

	l.logEngine(
		f.Envelope,
		"skipped",
		formatMessageType(mt),
		reason,
	)
}

// dispatchBegun logs a line that describes f.
func (l *Logger) dispatchBegun(f DispatchBegun) {
	l.Log(
		l.formatIDColumn(f.Envelope) + "  " +
			formatMessageType(message.TypeOf(f.Envelope.Message)),
	)
}

// commandDeduplicated logs a line that describes f.
func (l *Logger) commandDeduplicated(f CommandDeduplicated) {
	mt := message.TypeOf(f.Envelope.Message)

	l.logEngine(
		f.Envelope,
		"ignored",
		formatMessageType(mt),
		fmt.Sprintf("its idempotency key %q has already been used", f.Key),
	)
}

// handlingCompleted logs a line that describes f if the handler failed.
func (l *Logger) handlingCompleted(f HandlingCompleted) {
	if f.Error != nil {
		l.logHandler(
			f.Handler,
			"",
			"failed",
			f.Error.Error(),
		)
	}
}

// handlingSkipped logs a line that describes f.
func (l *Logger) handlingSkipped(f HandlingSkipped) {
	var reason string

	switch f.Reason {
	case HandlerTypeDisabled:
		reason = fmt.Sprintf("%s handlers are disabled", f.Handler.HandlerType())
	case IndividualHandlerDisabled:
		reason = "it is disabled during this tick of the test engine"
	case IndividualHandlerDisabledByConfiguration:
		reason = "it is disabled by its Configure() method"
	}

	l.logHandler(
		f.Handler,
		"",
		"skipped",
		reason,
	)
}

// tickCycleBegun logs a line that describes f.
func (l *Logger) tickCycleBegun(f TickCycleBegun) {
	l.logEngine(
		nil,
		"ticked",
		formatEngineTime(f.EngineTime),
		formatEnabledHandlers(f.EnabledHandlerTypes, f.EnabledHandlers),
	)
}

// tickCompleted logs a line that describes f if the handler failed.
func (l *Logger) tickCompleted(f TickCompleted) {
	if f.Error != nil {
		l.logHandler(
			f.Handler,
			"",
			"failed",
			"while ticking: "+f.Error.Error(),
		)
	}
}

// aggregateInstanceLoaded logs a line that describes f.
func (l *Logger) aggregateInstanceLoaded(f AggregateInstanceLoaded) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"loaded",
		f.Root.AggregateInstanceDescription(),
	)
}

// aggregateInstanceNotFound logs a line that describes f.
func (l *Logger) aggregateInstanceNotFound(f AggregateInstanceNotFound) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"not found",
	)
}

// aggregateInstanceCreated logs a line that describes f.
func (l *Logger) aggregateInstanceCreated(f AggregateInstanceCreated) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"created",
		f.Root.AggregateInstanceDescription(),
	)
}

// eventRecordedByAggregate logs a line that describes f.
func (l *Logger) eventRecordedByAggregate(f EventRecordedByAggregate) {
	mt := message.TypeOf(f.EventEnvelope.Message)

	l.logHandler(
		f.Handler,
		f.InstanceID,
		"recorded",
		f.Root.AggregateInstanceDescription(),
		mt.String(),
		f.EventEnvelope.Message.MessageDescription(),
	)
}

// messageLoggedByAggregate logs a line that describes f.
func (l *Logger) messageLoggedByAggregate(f MessageLoggedByAggregate) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"logged",
		f.Root.AggregateInstanceDescription(),
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// processInstanceLoaded logs a line that describes f.
func (l *Logger) processInstanceLoaded(f ProcessInstanceLoaded) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"loaded",
		f.Root.ProcessInstanceDescription(false),
	)
}

// processEventIgnored logs a line that describes f.
func (l *Logger) processEventIgnored(f ProcessEventIgnored) {
	l.logHandler(
		f.Handler,
		"",
		"ignored",
		"not routed to any instance",
	)
}

// processEventRoutedToEndedInstance logs a line that describes f.
func (l *Logger) processEventRoutedToEndedInstance(f ProcessEventRoutedToEndedInstance) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"ignored",
		"the instance has ended",
	)
}

// processDeadlineRoutedToEndedInstance logs a line that describes f.
func (l *Logger) processDeadlineRoutedToEndedInstance(f ProcessDeadlineRoutedToEndedInstance) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"ignored",
		"the instance has ended",
	)
}

// processInstanceNotFound logs a line that describes f.
func (l *Logger) processInstanceNotFound(f ProcessInstanceNotFound) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"not found",
	)
}

// processInstanceBegun logs a line that describes f.
func (l *Logger) processInstanceBegun(f ProcessInstanceBegun) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"begun",
		f.Root.ProcessInstanceDescription(false),
	)
}

// processInstanceEnded logs a line that describes f.
func (l *Logger) processInstanceEnded(f ProcessInstanceEnded) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"ended",
		f.Root.ProcessInstanceDescription(true),
	)
}

// commandExecutedByProcess logs a line that describes f.
func (l *Logger) commandExecutedByProcess(f CommandExecutedByProcess) {
	mt := message.TypeOf(f.CommandEnvelope.Message)

	l.logHandler(
		f.Handler,
		f.InstanceID,
		"executed",
		f.Root.ProcessInstanceDescription(false),
		mt.String(),
		f.CommandEnvelope.Message.MessageDescription(),
	)
}

// deadlineScheduledByProcess logs a line that describes f.
func (l *Logger) deadlineScheduledByProcess(f DeadlineScheduledByProcess) {
	mt := message.TypeOf(f.DeadlineEnvelope.Message)

	l.logHandler(
		f.Handler,
		f.InstanceID,
		"scheduled",
		f.Root.ProcessInstanceDescription(false),
		mt.String(),
		"due "+f.DeadlineEnvelope.ScheduledFor.Format(time.RFC3339),
		f.DeadlineEnvelope.Message.MessageDescription(),
	)
}

// messageLoggedByProcess logs a line that describes f.
func (l *Logger) messageLoggedByProcess(f MessageLoggedByProcess) {
	l.logHandler(
		f.Handler,
		f.InstanceID,
		"logged",
		f.Root.ProcessInstanceDescription(f.Ended),
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// eventRecordedByIntegration logs a line that describes f.
func (l *Logger) eventRecordedByIntegration(f EventRecordedByIntegration) {
	mt := message.TypeOf(f.EventEnvelope.Message)

	l.logHandler(
		f.Handler,
		"",
		"recorded",
		mt.String(),
		f.EventEnvelope.Message.MessageDescription(),
	)
}

// messageLoggedByIntegration logs a line that describes f.
func (l *Logger) messageLoggedByIntegration(f MessageLoggedByIntegration) {
	l.logHandler(
		f.Handler,
		"",
		"logged",
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// projectionCompactionCompleted logs a line that describes f.
func (l *Logger) projectionCompactionCompleted(f ProjectionCompactionCompleted) {
	if f.Error == nil {
		l.logHandler(
			f.Handler,
			"",
			"compacted",
		)
	} else {
		l.logHandler(
			f.Handler,
			"",
			"failed",
			"while compacting: "+f.Error.Error(),
		)
	}
}

// messageLoggedByProjection logs a line that describes f.
func (l *Logger) messageLoggedByProjection(f MessageLoggedByProjection) {
	l.logHandler(
		f.Handler,
		"",
		"logged",
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

const (
	// idsWidth is the width of the ID column when Trace is false. It is the
	// number of characters in an 8-digit message ID.
	idsWidth = 8

	// traceIDsWidth is the width of the ID column when Trace is true. It is
	// the number of characters in three 8-digit message IDs, with two spaces
	// after the first and " → " after the second.
	traceIDsWidth = 29

	// typeWidth is the width of the column of an indented line that shows the
	// handler's type. It is the number of characters in "integration", the
	// longest name of a handler type.
	typeWidth = 11

	// outcomeWidth is the width of the column of an indented line that says
	// what happened. It is the number of characters in the longest text that
	// the logger shows in that column, such as "scheduled".
	outcomeWidth = 9

	// separator is the text between the parts of the details.
	separator = " ● "
)

// logEngine logs a line that describes the engine's own work and is not a
// heading.
//
// env is the envelope of the message that the line is about, or nil if the
// line is not about a message. outcome is the word that says what happened.
// details is the parts of the details. Empty parts are left out.
func (l *Logger) logEngine(
	env *envelope.Envelope,
	outcome string,
	details ...string,
) {
	line := l.formatIDColumn(env) + "  " + outcome

	if d := joinDetails(details); d != "" {
		line += "  " + d
	}

	l.Log(line)
}

// logHandler logs an indented line that describes the work of handler h.
//
// instanceID is the ID of the instance that the line describes, or an empty
// string if it does not describe an instance. outcome is the word that says
// what happened. details is the parts of the details after the handler's name
// and the instance ID. Empty parts are left out.
func (l *Logger) logHandler(
	h config.Handler,
	instanceID string,
	outcome string,
	details ...string,
) {
	name := h.Identity().GetName()
	if instanceID != "" {
		name += " " + instanceID
	}

	l.Log(fmt.Sprintf(
		"%*s  %-*s  %-*s  %s",
		l.idsWidth(), "",
		typeWidth, h.HandlerType(),
		outcomeWidth, outcome,
		joinDetails(append([]string{name}, details...)),
	))
}

// joinDetails returns the non-empty parts of details, separated by separator.
func joinDetails(details []string) string {
	var parts []string
	for _, d := range details {
		if d != "" {
			parts = append(parts, d)
		}
	}
	return strings.Join(parts, separator)
}

// idsWidth returns the width of the ID column.
func (l *Logger) idsWidth() int {
	if l.Trace {
		return traceIDsWidth
	}
	return idsWidth
}

// formatIDColumn returns the ID column of a line about the message in env,
// padded to its width. It contains only spaces if env is nil.
func (l *Logger) formatIDColumn(env *envelope.Envelope) string {
	var ids string

	if env != nil {
		ids = formatMessageID(env.MessageID)

		if l.Trace && env.CausationID != env.MessageID {
			ids = formatMessageID(env.CorrelationID) + "  " +
				formatMessageID(env.CausationID) + " → " + ids
		}
	}

	return fmt.Sprintf("%*s", l.idsWidth(), ids)
}

// formatMessageType returns the name of mt without its package, followed by
// the kind of message, such as "OpenAccount command".
func formatMessageType(mt message.Type) string {
	name := strings.TrimLeft(mt.String(), "*")

	// Remove the package name, but not from any type arguments, such as in
	// "pkg.Generic[pkg.Arg]".
	end := strings.IndexByte(name, '[')
	if end == -1 {
		end = len(name)
	}
	if i := strings.LastIndexByte(name[:end], '.'); i != -1 {
		name = name[i+1:]
	}

	return name + " " + mt.Kind().String()
}

// formatMessageID returns the first 8 hexadecimal digits of id.
func formatMessageID(id uuid.UUID) string {
	return id.String()[:8]
}

func formatEngineTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatEnabledHandlers(
	byType map[config.HandlerType]bool,
	byName map[string]bool,
) string {
	var s strings.Builder

	for ht := range config.HandlerTypes() {
		if byType[ht] {
			s.WriteString(config.MapByHandlerType(
				ht,
				" +aggregates",
				" +processes",
				" +integrations",
				" +projections",
			))
		}
	}

	// sort the handler names to display them deterministically
	var sorted []string
	for n := range byName {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)

	for _, n := range sorted {
		if byName[n] {
			s.WriteString(" +" + n)
		} else {
			s.WriteString(" -" + n)
		}
	}

	return "enabled:" + s.String()
}
