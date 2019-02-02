package fact

import (
	"fmt"
	"time"

	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/logging"
	"github.com/dogmatiq/enginekit/message"
)

// Logger is an observer that logs human-readable messages to a log function.
type Logger struct {
	logger logging.Logger
}

// NewLogger returns a new logging observer that logs facts to using the given
// log function.
func NewLogger(log func(string)) *Logger {
	return &Logger{
		logger: logging.Logger{
			Log:             log,
			FormatMessageID: formatMessageID,
		},
	}
}

// Notify the observer of a fact.generates the log message for f.
func (l *Logger) Notify(f Fact) {
	switch x := f.(type) {
	case DispatchCycleBegun:
		l.dispatchCycleBegun(x)
	case DispatchCycleSkipped:
		l.dispatchCycleSkipped(x)
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
	case MessageLoggedByProjection:
		l.messageLoggedByProjection(x)
	}
}

// dispatchCycleBegun returns the log message for f.
func (l *Logger) dispatchCycleBegun(f DispatchCycleBegun) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
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

// dispatchCycleSkipped returns the log message for f.
func (l *Logger) dispatchCycleSkipped(f DispatchCycleSkipped) {
	l.logger.LogGeneric(
		message.Correlation{},
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
func (l *Logger) dispatchBegun(f DispatchBegun) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.SystemIcon,
			"",
		},
		message.TypeOf(f.Envelope.Message).String()+f.Envelope.Role.Marker(),
		message.ToString(f.Envelope.Message),
	)
}

// handlingCompleted returns the log message for f.
func (l *Logger) handlingCompleted(f HandlingCompleted) {
	if f.Error != nil {
		l.logger.LogGeneric(
			f.Envelope.Correlation,
			[]string{
				logging.InboundErrorIcon,
				logging.HandlerTypeIcon(f.HandlerType),
				logging.ErrorIcon,
			},
			f.HandlerName,
			f.Error.Error(),
		)
	}
}

// handlingSkipped returns the log message for f.
func (l *Logger) handlingSkipped(f HandlingSkipped) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.HandlerTypeIcon(f.HandlerType),
			"",
		},
		f.HandlerName,
		fmt.Sprintf(
			"handler skipped because %s handlers are disabled",
			f.HandlerType,
		),
	)
}

// tickCycleBegun returns the log message for f.
func (l *Logger) tickCycleBegun(f TickCycleBegun) {
	l.logger.LogGeneric(
		message.Correlation{},
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

// tickCompleted returns the log message for f.
func (l *Logger) tickCompleted(f TickCompleted) {
	if f.Error != nil {
		l.logger.LogGeneric(
			message.Correlation{},
			[]string{
				"",
				logging.HandlerTypeIcon(f.HandlerType),
				logging.ErrorIcon,
			},
			f.HandlerName,
			f.Error.Error(),
		)
	}
}

// aggregateInstanceLoaded returns the log message for f.
func (l *Logger) aggregateInstanceLoaded(f AggregateInstanceLoaded) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"loaded an existing instance",
	)
}

// aggregateInstanceNotFound returns the log message for f.
func (l *Logger) aggregateInstanceNotFound(f AggregateInstanceNotFound) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"instance does not yet exist",
	)
}

// aggregateInstanceCreated returns the log message for f.
func (l *Logger) aggregateInstanceCreated(f AggregateInstanceCreated) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"instance created",
	)
}

// aggregateInstanceDestroyed returns the log message for f.
func (l *Logger) aggregateInstanceDestroyed(f AggregateInstanceDestroyed) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"instance destroyed",
	)
}

// eventRecordedByAggregate returns the log message for f.
func (l *Logger) eventRecordedByAggregate(f EventRecordedByAggregate) {
	l.logger.LogGeneric(
		f.EventEnvelope.Correlation,
		[]string{
			logging.OutboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"recorded an event",
		f.EventEnvelope.Type.String()+f.EventEnvelope.Role.Marker(),
		message.ToString(f.EventEnvelope.Message),
	)
}

// messageLoggedByAggregate returns the log message for f.
func (l *Logger) messageLoggedByAggregate(f MessageLoggedByAggregate) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.AggregateIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// processInstanceLoaded returns the log message for f.
func (l *Logger) processInstanceLoaded(f ProcessInstanceLoaded) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"loaded an existing instance",
	)
}

// processEventIgnored returns the log message for f.
func (l *Logger) processEventIgnored(f ProcessEventIgnored) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName,
		"event ignored because it was not routed to any instance",
	)
}

// processTimeoutIgnored returns the log message for f.
func (l *Logger) processTimeoutIgnored(f ProcessTimeoutIgnored) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"timeout ignored because the target instance no longer exists",
	)
}

// processInstanceNotFound returns the log message for f.
func (l *Logger) processInstanceNotFound(f ProcessInstanceNotFound) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"instance does not yet exist",
	)
}

// processInstanceBegun returns the log message for f.
func (l *Logger) processInstanceBegun(f ProcessInstanceBegun) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"instance begun",
	)
}

// processInstanceEnded returns the log message for f.
func (l *Logger) processInstanceEnded(f ProcessInstanceEnded) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"instance ended",
	)
}

// commandExecutedByProcess returns the log message for f.
func (l *Logger) commandExecutedByProcess(f CommandExecutedByProcess) {
	l.logger.LogGeneric(
		f.CommandEnvelope.Correlation,
		[]string{
			logging.OutboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		"executed a command",
		f.CommandEnvelope.Type.String()+f.CommandEnvelope.Role.Marker(),
		message.ToString(f.CommandEnvelope.Message),
	)
}

// timeoutScheduledByProcess returns the log message for f.
func (l *Logger) timeoutScheduledByProcess(f TimeoutScheduledByProcess) {
	l.logger.LogGeneric(
		f.TimeoutEnvelope.Correlation,
		[]string{
			logging.OutboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		fmt.Sprintf(
			"scheduled a timeout for %s",
			f.TimeoutEnvelope.TimeoutTime.Format(time.RFC3339),
		),
		f.TimeoutEnvelope.Type.String()+f.TimeoutEnvelope.Role.Marker(),
		message.ToString(f.TimeoutEnvelope.Message),
	)
}

// messageLoggedByProcess returns the log message for f.
func (l *Logger) messageLoggedByProcess(f MessageLoggedByProcess) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProcessIcon,
			"",
		},
		f.HandlerName+" "+f.InstanceID,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// eventRecordedByIntegration returns the log message for f.
func (l *Logger) eventRecordedByIntegration(f EventRecordedByIntegration) {
	l.logger.LogGeneric(
		f.EventEnvelope.Correlation,
		[]string{
			logging.OutboundIcon,
			logging.IntegrationIcon,
			"",
		},
		f.HandlerName,
		"recorded an event",
		f.EventEnvelope.Type.String()+f.EventEnvelope.Role.Marker(),
		message.ToString(f.EventEnvelope.Message),
	)
}

// messageLoggedByIntegration returns the log message for f.
func (l *Logger) messageLoggedByIntegration(f MessageLoggedByIntegration) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.IntegrationIcon,
			"",
		},
		f.HandlerName,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// messageLoggedByProjection returns the log message for f.
func (l *Logger) messageLoggedByProjection(f MessageLoggedByProjection) {
	l.logger.LogGeneric(
		f.Envelope.Correlation,
		[]string{
			logging.InboundIcon,
			logging.ProjectionIcon,
			"",
		},
		f.HandlerName,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

func formatMessageID(id string) string {
	if id == "" {
		return "----"
	}

	return fmt.Sprintf("%04s", id)
}

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
