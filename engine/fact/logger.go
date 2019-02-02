package fact

import (
	"fmt"
	"strings"
	"time"

	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/logging"
	"github.com/dogmatiq/enginekit/message"
)

const truncateLength = 4

var emptyCorrelation = message.NewCorrelation(
	strings.Repeat("-", truncateLength),
)

func formatEnabledHandlers(e map[handler.Type]bool) string {
	var s string

	for _, t := range handler.Types {
		if e[t] {
			if s != "" {
				s += ", "
			}

			s += t.String()
		}
	}

	return s
}

func formatHandler(n, s string) string {
	return "[" + n + "]  " + s
}

func formatHandlerAndInstance(n, id, s string) string {
	return "[" + n + " " + id + "]  " + s
}

// Logger is an observer that logs human-readable messages to a logger.
type Logger struct {
	Log func(string, ...interface{})
}

// Notify the observer of a fact.generates the log message for f.
func (l *Logger) Notify(f Fact) {
	var m string

	switch x := f.(type) {
	case DispatchCycleBegun:
		m = l.dispatchCycleBegun(x)
	case DispatchCycleCompleted:
		m = l.dispatchCycleCompleted(x)
	case DispatchCycleSkipped:
		m = l.dispatchCycleSkipped(x)
	case DispatchBegun:
		m = l.dispatchBegun(x)
	case DispatchCompleted:
		m = l.dispatchCompleted(x)
	case HandlingBegun:
		m = l.handlingBegun(x)
	case HandlingCompleted:
		m = l.handlingCompleted(x)
	case HandlingSkipped:
		m = l.handlingSkipped(x)
	case TickCycleBegun:
		m = l.tickCycleBegun(x)
	case TickCycleCompleted:
		m = l.tickCycleCompleted(x)
	case TickBegun:
		m = l.tickBegun(x)
	case TickCompleted:
		m = l.tickCompleted(x)
	case AggregateInstanceLoaded:
		m = l.aggregateInstanceLoaded(x)
	case AggregateInstanceNotFound:
		m = l.aggregateInstanceNotFound(x)
	case AggregateInstanceCreated:
		m = l.aggregateInstanceCreated(x)
	case AggregateInstanceDestroyed:
		m = l.aggregateInstanceDestroyed(x)
	case EventRecordedByAggregate:
		m = l.eventRecordedByAggregate(x)
	case MessageLoggedByAggregate:
		m = l.messageLoggedByAggregate(x)
	case ProcessInstanceLoaded:
		m = l.processInstanceLoaded(x)
	case ProcessEventIgnored:
		m = l.processEventIgnored(x)
	case ProcessTimeoutIgnored:
		m = l.processTimeoutIgnored(x)
	case ProcessInstanceNotFound:
		m = l.processInstanceNotFound(x)
	case ProcessInstanceBegun:
		m = l.processInstanceBegun(x)
	case ProcessInstanceEnded:
		m = l.processInstanceEnded(x)
	case CommandExecutedByProcess:
		m = l.commandExecutedByProcess(x)
	case TimeoutScheduledByProcess:
		m = l.timeoutScheduledByProcess(x)
	case MessageLoggedByProcess:
		m = l.messageLoggedByProcess(x)
	case EventRecordedByIntegration:
		m = l.eventRecordedByIntegration(x)
	case MessageLoggedByIntegration:
		m = l.messageLoggedByIntegration(x)
	case MessageLoggedByProjection:
		m = l.messageLoggedByProjection(x)
	default:
		panic("unrecognised fact")
	}

	if m != "" && l.Log != nil {
		l.Log(m)
	}
}

// dispatchCycleBegun returns the log message for f.
func (l *Logger) dispatchCycleBegun(f DispatchCycleBegun) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.SystemIcon,
			"",
		},
		fmt.Sprintf(
			"dispatch cycle begun at %s [enabled: %s]",
			f.EngineTime.Format(time.RFC3339),
			formatEnabledHandlers(f.EnabledHandlers),
		),
	)
}

// dispatchCycleCompleted returns the log message for f.
func (l *Logger) dispatchCycleCompleted(f DispatchCycleCompleted) string {
	if f.Error == nil {
		return logging.Format(
			f.Envelope.Correlation,
			truncateLength,
			[]string{
				logging.InboundIcon,
				logging.SystemIcon,
				"",
			},
			"dispatch cycle completed successfully",
		)
	}

	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundErrorIcon,
			logging.SystemIcon,
			logging.ErrorIcon,
		},
		"dispatch cycle completed with errors",
	)
}

// dispatchCycleSkipped returns the log message for f.
func (l *Logger) dispatchCycleSkipped(f DispatchCycleSkipped) string {
	return logging.Format(
		emptyCorrelation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.SystemIcon,
			"",
		},
		message.TypeOf(f.Message).String(),
		"dispatch cycle skipped because this message type is not routed to any handlers",
	)
}

// dispatchBegun returns the log message for f.
func (l *Logger) dispatchBegun(f DispatchBegun) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.SystemIcon,
			"",
		},
		message.TypeOf(f.Envelope.Message).String()+f.Envelope.Role.Marker(),
		message.ToString(f.Envelope.Message),
		"dispatch begun",
	)
}

// dispatchCompleted returns the log message for f.
func (l *Logger) dispatchCompleted(f DispatchCompleted) string {
	if f.Error == nil {
		return logging.Format(
			f.Envelope.Correlation,
			truncateLength,
			[]string{
				logging.InboundIcon,
				logging.SystemIcon,
				"",
			},
			"dispatch completed successfully",
		)
	}

	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundErrorIcon,
			logging.SystemIcon,
			logging.ErrorIcon,
		},
		"dispatch completed with errors",
	)
}

// handlingBegun returns the log message for f.
func (l *Logger) handlingBegun(f HandlingBegun) string {
	return ""
}

// handlingCompleted returns the log message for f.
func (l *Logger) handlingCompleted(f HandlingCompleted) string {
	if f.Error == nil {
		return ""
	}

	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundErrorIcon,
			logging.HandlerTypeIcon(f.HandlerType),
			logging.ErrorIcon,
		},
		formatHandler(
			f.HandlerName,
			f.Error.Error(),
		),
	)
}

// handlingSkipped returns the log message for f.
func (l *Logger) handlingSkipped(f HandlingSkipped) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.HandlerTypeIcon(f.HandlerType),
			"",
		},
		formatHandler(
			f.HandlerName,
			fmt.Sprintf(
				"handler skipped because %s handlers are disabled",
				f.HandlerType,
			),
		),
	)
}

// tickCycleBegun returns the log message for f.
func (l *Logger) tickCycleBegun(f TickCycleBegun) string {
	return logging.Format(
		emptyCorrelation,
		truncateLength,
		[]string{
			"",
			logging.SystemIcon,
			"",
		},
		fmt.Sprintf(
			"tick cycle begun at %s [enabled: %s]",
			f.EngineTime.Format(time.RFC3339),
			formatEnabledHandlers(f.EnabledHandlers),
		),
	)
}

// tickCycleCompleted  returns the log message for f.
func (l *Logger) tickCycleCompleted(f TickCycleCompleted) string {
	if f.Error == nil {
		return logging.Format(
			emptyCorrelation,
			truncateLength,
			[]string{
				"",
				logging.SystemIcon,
				"",
			},
			"tick cycle completed successfully",
		)
	}

	return logging.Format(
		emptyCorrelation,
		truncateLength,
		[]string{
			"",
			logging.SystemIcon,
			logging.ErrorIcon,
		},
		"tick cycle completed with errors",
	)
}

// tickBegun returns the log message for f.
func (l *Logger) tickBegun(f TickBegun) string {
	return ""
}

// tickCompleted returns the log message for f.
func (l *Logger) tickCompleted(f TickCompleted) string {
	if f.Error == nil {
		return ""
	}

	return logging.Format(
		emptyCorrelation,
		truncateLength,
		[]string{
			"",
			logging.HandlerTypeIcon(f.HandlerType),
			logging.ErrorIcon,
		},
		formatHandler(
			f.HandlerName,
			f.Error.Error(),
		),
	)
}

// aggregateInstanceLoaded returns the log message for f.
func (l *Logger) aggregateInstanceLoaded(f AggregateInstanceLoaded) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"loaded an existing instance",
		),
	)
}

// aggregateInstanceNotFound returns the log message for f.
func (l *Logger) aggregateInstanceNotFound(f AggregateInstanceNotFound) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"instance does not yet exist",
		),
	)
}

// aggregateInstanceCreated returns the log message for f.
func (l *Logger) aggregateInstanceCreated(f AggregateInstanceCreated) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"instance created",
		),
	)
}

// aggregateInstanceDestroyed returns the log message for f.
func (l *Logger) aggregateInstanceDestroyed(f AggregateInstanceDestroyed) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"instance destroyed",
		),
	)
}

// eventRecordedByAggregate returns the log message for f.
func (l *Logger) eventRecordedByAggregate(f EventRecordedByAggregate) string {
	return logging.Format(
		f.EventEnvelope.Correlation,
		truncateLength,
		[]string{
			logging.OutboundIcon,
			logging.AggregateIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"recorded an event",
		),
		f.EventEnvelope.Type.String()+f.EventEnvelope.Role.Marker(),
		message.ToString(f.EventEnvelope.Message),
	)
}

// messageLoggedByAggregate returns the log message for f.
func (l *Logger) messageLoggedByAggregate(f MessageLoggedByAggregate) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			fmt.Sprintf(f.LogFormat, f.LogArguments...),
		),
	)
}

// processInstanceLoaded returns the log message for f.
func (l *Logger) processInstanceLoaded(f ProcessInstanceLoaded) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"loaded an existing instance",
		),
	)
}

// processEventIgnored returns the log message for f.
func (l *Logger) processEventIgnored(f ProcessEventIgnored) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandler(
			f.HandlerName,
			"event ignored because it was not routed to any instance",
		),
	)
}

// processTimeoutIgnored returns the log message for f.
func (l *Logger) processTimeoutIgnored(f ProcessTimeoutIgnored) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"timeout ignored because the target instance no longer exists",
		),
	)
}

// processInstanceNotFound returns the log message for f.
func (l *Logger) processInstanceNotFound(f ProcessInstanceNotFound) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"instance does not yet exist",
		),
	)
}

// processInstanceBegun returns the log message for f.
func (l *Logger) processInstanceBegun(f ProcessInstanceBegun) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"instance begun",
		),
	)
}

// processInstanceEnded returns the log message for f.
func (l *Logger) processInstanceEnded(f ProcessInstanceEnded) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"instance ended",
		),
	)
}

// commandExecutedByProcess returns the log message for f.
func (l *Logger) commandExecutedByProcess(f CommandExecutedByProcess) string {
	return logging.Format(
		f.CommandEnvelope.Correlation,
		truncateLength,
		[]string{
			logging.OutboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			"executed a command",
		),
		f.CommandEnvelope.Type.String()+f.CommandEnvelope.Role.Marker(),
		message.ToString(f.CommandEnvelope.Message),
	)
}

// timeoutScheduledByProcess returns the log message for f.
func (l *Logger) timeoutScheduledByProcess(f TimeoutScheduledByProcess) string {
	return logging.Format(
		f.TimeoutEnvelope.Correlation,
		truncateLength,
		[]string{
			logging.OutboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			fmt.Sprintf(
				"scheduled a timeout for %s",
				f.TimeoutEnvelope.TimeoutTime.Format(time.RFC3339),
			),
		),
		f.TimeoutEnvelope.Type.String()+f.TimeoutEnvelope.Role.Marker(),
		message.ToString(f.TimeoutEnvelope.Message),
	)
}

// messageLoggedByProcess returns the log message for f.
func (l *Logger) messageLoggedByProcess(f MessageLoggedByProcess) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		formatHandlerAndInstance(
			f.HandlerName,
			f.InstanceID,
			fmt.Sprintf(f.LogFormat, f.LogArguments...),
		),
	)
}

// eventRecordedByIntegration returns the log message for f.
func (l *Logger) eventRecordedByIntegration(f EventRecordedByIntegration) string {
	return logging.Format(
		f.EventEnvelope.Correlation,
		truncateLength,
		[]string{
			logging.OutboundIcon,
			logging.IntegrationIcon,
			"",
		},
		formatHandler(
			f.HandlerName,
			"recorded an event",
		),
		f.EventEnvelope.Type.String()+f.EventEnvelope.Role.Marker(),
		message.ToString(f.EventEnvelope.Message),
	)
}

// messageLoggedByIntegration returns the log message for f.
func (l *Logger) messageLoggedByIntegration(f MessageLoggedByIntegration) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.IntegrationIcon,
			"",
		},
		formatHandler(
			f.HandlerName,
			fmt.Sprintf(f.LogFormat, f.LogArguments...),
		),
	)
}

// messageLoggedByProjection returns the log message for f.
func (l *Logger) messageLoggedByProjection(f MessageLoggedByProjection) string {
	return logging.Format(
		f.Envelope.Correlation,
		truncateLength,
		[]string{
			logging.InboundIcon,
			logging.ProjectionIcon,
			"",
		},
		formatHandler(
			f.HandlerName,
			fmt.Sprintf(f.LogFormat, f.LogArguments...),
		),
	)
}
