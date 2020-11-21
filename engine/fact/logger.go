package fact

import (
	"fmt"
	"sort"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine/envelope"
	"github.com/dogmatiq/testkit/engine/fact/internal/logging"
)

// Logger is an observer that logs human-readable messages to a log function.
type Logger struct {
	Log func(string)
}

// NewLogger returns a new observer that logs human-readable descriptions of
// facts to the given log function.
func NewLogger(log func(string)) *Logger {
	return &Logger{
		Log: log,
	}
}

// Notify the observer of a fact.generates the log message for f.
func (l *Logger) Notify(f Fact) {
	switch x := f.(type) {
	case DispatchCycleBegun:
		l.dispatchCycleBegun(x)
	case DispatchBegun:
		l.dispatchBegun(x)
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
	case AggregateInstanceDestroyed:
		l.aggregateInstanceDestroyed(x)
	case EventRecordedByAggregate:
		l.eventRecordedByAggregate(x)
	case MessageLoggedByAggregate:
		l.messageLoggedByAggregate(x)
	case ProcessInstanceLoaded:
		l.processInstanceLoaded(x)
	case ProcessEventIgnored:
		l.processEventIgnored(x)
	case ProcessTimeoutIgnored:
		l.processTimeoutIgnored(x)
	case ProcessInstanceNotFound:
		l.processInstanceNotFound(x)
	case ProcessInstanceBegun:
		l.processInstanceBegun(x)
	case ProcessInstanceEnded:
		l.processInstanceEnded(x)
	case CommandExecutedByProcess:
		l.commandExecutedByProcess(x)
	case TimeoutScheduledByProcess:
		l.timeoutScheduledByProcess(x)
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

// dispatchCycleBegun returns the log message for f.
func (l *Logger) dispatchCycleBegun(f DispatchCycleBegun) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.SystemIcon,
			"",
		},
		"dispatching",
		formatEngineTime(f.EngineTime),
		formatEnabledHandlers(f.EnabledHandlerTypes, f.EnabledHandlers),
	)
}

// dispatchBegun returns the log message for f.
func (l *Logger) dispatchBegun(f DispatchBegun) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.SystemIcon,
			"",
		},
		message.TypeOf(f.Envelope.Message).String()+f.Envelope.Role.Marker(),
		dogma.DescribeMessage(f.Envelope.Message),
	)
}

// handlingCompleted returns the log message for f.
func (l *Logger) handlingCompleted(f HandlingCompleted) {
	if f.Error != nil {
		l.log(
			f.Envelope,
			[]logging.Icon{
				logging.InboundErrorIcon,
				logging.HandlerTypeIcon(f.Handler.HandlerType()),
				logging.ErrorIcon,
			},
			f.Handler.Identity().Name,
			f.Error.Error(),
		)
	}
}

// handlingSkipped returns the log message for f.
func (l *Logger) handlingSkipped(f HandlingSkipped) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.HandlerTypeIcon(f.Handler.HandlerType()),
			"",
		},
		f.Handler.Identity().Name,
		fmt.Sprintf(
			"handler skipped because %s handlers are disabled",
			f.Handler.HandlerType(),
		),
	)
}

// tickCycleBegun returns the log message for f.
func (l *Logger) tickCycleBegun(f TickCycleBegun) {
	l.log(
		&envelope.Envelope{},
		[]logging.Icon{
			"",
			logging.SystemIcon,
			"",
		},
		"ticking",
		formatEngineTime(f.EngineTime),
		formatEnabledHandlers(f.EnabledHandlerTypes, f.EnabledHandlers),
	)
}

// tickCompleted returns the log message for f.
func (l *Logger) tickCompleted(f TickCompleted) {
	if f.Error != nil {
		l.log(
			&envelope.Envelope{},
			[]logging.Icon{
				"",
				logging.HandlerTypeIcon(f.Handler.HandlerType()),
				logging.ErrorIcon,
			},
			f.Handler.Identity().Name,
			f.Error.Error(),
		)
	}
}

// aggregateInstanceLoaded returns the log message for f.
func (l *Logger) aggregateInstanceLoaded(f AggregateInstanceLoaded) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"loaded an existing instance",
	)
}

// aggregateInstanceNotFound returns the log message for f.
func (l *Logger) aggregateInstanceNotFound(f AggregateInstanceNotFound) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"instance does not yet exist",
	)
}

// aggregateInstanceCreated returns the log message for f.
func (l *Logger) aggregateInstanceCreated(f AggregateInstanceCreated) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"instance created",
	)
}

// aggregateInstanceDestroyed returns the log message for f.
func (l *Logger) aggregateInstanceDestroyed(f AggregateInstanceDestroyed) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"instance destroyed",
	)
}

// eventRecordedByAggregate returns the log message for f.
func (l *Logger) eventRecordedByAggregate(f EventRecordedByAggregate) {
	l.log(
		f.EventEnvelope,
		[]logging.Icon{
			logging.OutboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"recorded an event",
		f.EventEnvelope.Type.String()+f.EventEnvelope.Role.Marker(),
		dogma.DescribeMessage(f.EventEnvelope.Message),
	)
}

// messageLoggedByAggregate returns the log message for f.
func (l *Logger) messageLoggedByAggregate(f MessageLoggedByAggregate) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// processInstanceLoaded returns the log message for f.
func (l *Logger) processInstanceLoaded(f ProcessInstanceLoaded) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"loaded an existing instance",
	)
}

// processEventIgnored returns the log message for f.
func (l *Logger) processEventIgnored(f ProcessEventIgnored) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name,
		"event ignored because it was not routed to any instance",
	)
}

// processTimeoutIgnored returns the log message for f.
func (l *Logger) processTimeoutIgnored(f ProcessTimeoutIgnored) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"timeout ignored because the target instance no longer exists",
	)
}

// processInstanceNotFound returns the log message for f.
func (l *Logger) processInstanceNotFound(f ProcessInstanceNotFound) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"instance does not yet exist",
	)
}

// processInstanceBegun returns the log message for f.
func (l *Logger) processInstanceBegun(f ProcessInstanceBegun) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"instance begun",
	)
}

// processInstanceEnded returns the log message for f.
func (l *Logger) processInstanceEnded(f ProcessInstanceEnded) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"instance ended",
	)
}

// commandExecutedByProcess returns the log message for f.
func (l *Logger) commandExecutedByProcess(f CommandExecutedByProcess) {
	l.log(
		f.CommandEnvelope,
		[]logging.Icon{
			logging.OutboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		"executed a command",
		f.CommandEnvelope.Type.String()+f.CommandEnvelope.Role.Marker(),
		dogma.DescribeMessage(f.CommandEnvelope.Message),
	)
}

// timeoutScheduledByProcess returns the log message for f.
func (l *Logger) timeoutScheduledByProcess(f TimeoutScheduledByProcess) {
	l.log(
		f.TimeoutEnvelope,
		[]logging.Icon{
			logging.OutboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		fmt.Sprintf(
			"scheduled a timeout for %s",
			f.TimeoutEnvelope.ScheduledFor.Format(time.RFC3339),
		),
		f.TimeoutEnvelope.Type.String()+f.TimeoutEnvelope.Role.Marker(),
		dogma.DescribeMessage(f.TimeoutEnvelope.Message),
	)
}

// messageLoggedByProcess returns the log message for f.
func (l *Logger) messageLoggedByProcess(f MessageLoggedByProcess) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.Handler.Identity().Name+" "+f.InstanceID,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// eventRecordedByIntegration returns the log message for f.
func (l *Logger) eventRecordedByIntegration(f EventRecordedByIntegration) {
	l.log(
		f.EventEnvelope,
		[]logging.Icon{
			logging.OutboundIcon,
			logging.IntegrationIcon,
			"",
		},
		f.Handler.Identity().Name,
		"recorded an event",
		f.EventEnvelope.Type.String()+f.EventEnvelope.Role.Marker(),
		dogma.DescribeMessage(f.EventEnvelope.Message),
	)
}

// messageLoggedByIntegration returns the log message for f.
func (l *Logger) messageLoggedByIntegration(f MessageLoggedByIntegration) {
	l.log(
		f.Envelope,
		[]logging.Icon{
			logging.InboundIcon,
			logging.IntegrationIcon,
			"",
		},
		f.Handler.Identity().Name,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// projectionCompactionCompleted returns the log message for f.
func (l *Logger) projectionCompactionCompleted(f ProjectionCompactionCompleted) {
	if f.Error == nil {
		l.log(
			nil,
			[]logging.Icon{
				"",
				logging.ProjectionIcon,
				"",
			},
			f.Handler.Identity().Name,
			"compacted",
		)
	} else {
		l.log(
			nil,
			[]logging.Icon{
				"",
				logging.ProjectionIcon,
				logging.ErrorIcon,
			},
			f.Handler.Identity().Name,
			fmt.Sprintf("compaction failed: %s", f.Error),
		)
	}
}

// messageLoggedByProjection returns the log message for f.
func (l *Logger) messageLoggedByProjection(f MessageLoggedByProjection) {
	icons := []logging.Icon{
		"",
		logging.ProjectionIcon,
		"",
	}

	if f.Envelope != nil {
		icons[0] = logging.InboundIcon
	}

	l.log(
		f.Envelope,
		icons,
		f.Handler.Identity().Name,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

func (l *Logger) log(
	env *envelope.Envelope,
	icons []logging.Icon,
	text ...string,
) {
	var messageID, causationID, correlationID string
	if env != nil {
		messageID = env.MessageID
		causationID = env.CausationID
		correlationID = env.CorrelationID
	}

	l.Log(logging.String(
		[]logging.IconWithLabel{
			logging.MessageIDIcon.WithLabel(
				formatMessageID(messageID),
			),
			logging.CausationIDIcon.WithLabel(
				formatMessageID(causationID),
			),
			logging.CorrelationIDIcon.WithLabel(
				formatMessageID(correlationID),
			),
		},
		icons,
		text...,
	))
}

func formatMessageID(id string) string {
	if id == "" {
		return "----"
	}

	return fmt.Sprintf("%04s", id)
}

func formatEngineTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

var handlerTypePlurals = map[configkit.HandlerType]string{
	configkit.AggregateHandlerType:   "aggregates",
	configkit.ProcessHandlerType:     "processes",
	configkit.IntegrationHandlerType: "integrations",
	configkit.ProjectionHandlerType:  "projections",
}

func formatEnabledHandlers(
	byType map[configkit.HandlerType]bool,
	byName map[string]bool,
) string {
	var s string

	for _, t := range configkit.HandlerTypes {
		if byType[t] {
			s += " +" + handlerTypePlurals[t]
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
			s += " +" + n
		} else {
			s += " -" + n
		}
	}

	return "enabled:" + s
}
