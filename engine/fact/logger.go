package fact

import (
	"fmt"
	"time"

	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
)

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

	if l.Log != nil {
		l.Log(m)
	}
}

// formatEnabledHandlers returns a list of the enabled handler types as a string.
func (l *Logger) formatEnabledHandlers(e map[handler.Type]bool) string {
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

// dispatchCycleBegun returns the log message for f.
func (l *Logger) dispatchCycleBegun(f DispatchCycleBegun) string {
	return fmt.Sprintf(
		"engine: dispatch of '%s' %s begun at %s (enabled: %s)",
		f.Envelope.Type,
		f.Envelope.Role,
		f.Now.Format(time.RFC3339),
		l.formatEnabledHandlers(f.EnabledHandlers),
	)
}

// dispatchCycleCompleted returns the log message for f.
func (l *Logger) dispatchCycleCompleted(f DispatchCycleCompleted) string {
	if f.Error == nil {
		return fmt.Sprintf(
			"engine: dispatch of '%s' %s completed successfully",
			f.Envelope.Type,
			f.Envelope.Role,
		)
	}

	return fmt.Sprintf(
		"engine: dispatch of '%s' %s completed with errors",
		f.Envelope.Type,
		f.Envelope.Role,
	)
}

// dispatchCycleSkipped returns the log message for f.
func (l *Logger) dispatchCycleSkipped(f DispatchCycleSkipped) string {
	return fmt.Sprintf(
		"engine: no route for '%s' messages",
		message.TypeOf(f.Message),
	)
}

// handlingBegun returns the log message for f.
func (l *Logger) handlingBegun(f HandlingBegun) string {
	return fmt.Sprintf(
		"%s[%s]: message handling begun",
		f.HandlerType,
		f.HandlerName,
	)
}

// handlingCompleted returns the log message for f.
func (l *Logger) handlingCompleted(f HandlingCompleted) string {
	if f.Error == nil {
		return fmt.Sprintf(
			"%s[%s]: handled message successfully",
			f.HandlerType,
			f.HandlerName,
		)
	}

	return fmt.Sprintf(
		"%s[%s]: handling failed: %s",
		f.HandlerType,
		f.HandlerName,
		f.Error,
	)
}

// handlingSkipped returns the log message for f.
func (l *Logger) handlingSkipped(f HandlingSkipped) string {
	return fmt.Sprintf(
		"%s[%s]: message handling skipped because %s handlers are disabled",
		f.HandlerType,
		f.HandlerName,
		f.HandlerType,
	)
}

// tickCycleBegun returns the log message for f.
func (l *Logger) tickCycleBegun(f TickCycleBegun) string {
	return fmt.Sprintf(
		"engine: tick begun at %s (enabled: %s)",
		f.Now.Format(time.RFC3339),
		l.formatEnabledHandlers(f.EnabledHandlers),
	)
}

// tickCycleCompleted  returns the log message for f.
func (l *Logger) tickCycleCompleted(f TickCycleCompleted) string {
	if f.Error == nil {
		return "engine: tick completed successfully"
	}

	return "engine: tick completed with errors"
}

// tickBegun returns the log message for f.
func (l *Logger) tickBegun(f TickBegun) string {
	return fmt.Sprintf(
		"%s[%s]: tick begun",
		f.HandlerType,
		f.HandlerName,
	)
}

// tickCompleted returns the log message for f.
func (l *Logger) tickCompleted(f TickCompleted) string {
	if f.Error == nil {
		return fmt.Sprintf(
			"%s[%s]: tick completed successfully",
			f.HandlerType,
			f.HandlerName,
		)
	}

	return fmt.Sprintf(
		"%s[%s]: tick failed: %s",
		f.HandlerType,
		f.HandlerName,
		f.Error,
	)
}

// aggregateInstanceLoaded returns the log message for f.
func (l *Logger) aggregateInstanceLoaded(f AggregateInstanceLoaded) string {
	return fmt.Sprintf(
		"aggregate[%s@%s]: loading existing instance",
		f.HandlerName,
		f.InstanceID,
	)
}

// aggregateInstanceNotFound returns the log message for f.
func (l *Logger) aggregateInstanceNotFound(f AggregateInstanceNotFound) string {
	return fmt.Sprintf(
		"aggregate[%s@%s]: no existing instance found",
		f.HandlerName,
		f.InstanceID,
	)
}

// aggregateInstanceCreated returns the log message for f.
func (l *Logger) aggregateInstanceCreated(f AggregateInstanceCreated) string {
	return fmt.Sprintf(
		"aggregate[%s@%s]: instance created",
		f.HandlerName,
		f.InstanceID,
	)
}

// aggregateInstanceDestroyed returns the log message for f.
func (l *Logger) aggregateInstanceDestroyed(f AggregateInstanceDestroyed) string {
	return fmt.Sprintf(
		"aggregate[%s@%s]: instance destroyed",
		f.HandlerName,
		f.InstanceID,
	)
}

// eventRecordedByAggregate returns the log message for f.
func (l *Logger) eventRecordedByAggregate(f EventRecordedByAggregate) string {
	return fmt.Sprintf(
		"aggregate[%s@%s]: recorded '%s' event",
		f.HandlerName,
		f.InstanceID,
		f.EventEnvelope.Type,
	)
}

// messageLoggedByAggregate returns the log message for f.
func (l *Logger) messageLoggedByAggregate(f MessageLoggedByAggregate) string {
	return fmt.Sprintf(
		"aggregate[%s@%s]: %s",
		f.HandlerName,
		f.InstanceID,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// processInstanceLoaded returns the log message for f.
func (l *Logger) processInstanceLoaded(f ProcessInstanceLoaded) string {
	return fmt.Sprintf(
		"process[%s@%s]: loading existing instance",
		f.HandlerName,
		f.InstanceID,
	)
}

// processEventIgnored returns the log message for f.
func (l *Logger) processEventIgnored(f ProcessEventIgnored) string {
	return fmt.Sprintf(
		"process[%s]: event not routed to any instance",
		f.HandlerName,
	)
}

// processTimeoutIgnored returns the log message for f.
func (l *Logger) processTimeoutIgnored(f ProcessTimeoutIgnored) string {
	return fmt.Sprintf(
		"process[%s@%s]: timeout's instance no longer exists",
		f.HandlerName,
		f.InstanceID,
	)
}

// processInstanceNotFound returns the log message for f.
func (l *Logger) processInstanceNotFound(f ProcessInstanceNotFound) string {
	return fmt.Sprintf(
		"process[%s@%s]: no existing instance found",
		f.HandlerName,
		f.InstanceID,
	)
}

// processInstanceBegun returns the log message for f.
func (l *Logger) processInstanceBegun(f ProcessInstanceBegun) string {
	return fmt.Sprintf(
		"process[%s@%s]: instance begun",
		f.HandlerName,
		f.InstanceID,
	)
}

// processInstanceEnded returns the log message for f.
func (l *Logger) processInstanceEnded(f ProcessInstanceEnded) string {
	return fmt.Sprintf(
		"process[%s@%s]: instance ended",
		f.HandlerName,
		f.InstanceID,
	)
}

// commandExecutedByProcess returns the log message for f.
func (l *Logger) commandExecutedByProcess(f CommandExecutedByProcess) string {
	return fmt.Sprintf(
		"process[%s@%s]: executed '%s' command",
		f.HandlerName,
		f.InstanceID,
		f.CommandEnvelope.Type,
	)
}

// timeoutScheduledByProcess returns the log message for f.
func (l *Logger) timeoutScheduledByProcess(f TimeoutScheduledByProcess) string {
	return fmt.Sprintf(
		"process[%s@%s]: scheduled '%s' timeout for %s",
		f.HandlerName,
		f.InstanceID,
		f.TimeoutEnvelope.Type,
		f.TimeoutEnvelope.TimeoutTime.Format(time.RFC3339),
	)
}

// messageLoggedByProcess returns the log message for f.
func (l *Logger) messageLoggedByProcess(f MessageLoggedByProcess) string {
	return fmt.Sprintf(
		"process[%s@%s]: %s",
		f.HandlerName,
		f.InstanceID,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// eventRecordedByIntegration returns the log message for f.
func (l *Logger) eventRecordedByIntegration(f EventRecordedByIntegration) string {
	return fmt.Sprintf(
		"integration[%s]: recorded '%s' event",
		f.HandlerName,
		f.EventEnvelope.Type,
	)
}

// messageLoggedByIntegration returns the log message for f.
func (l *Logger) messageLoggedByIntegration(f MessageLoggedByIntegration) string {
	return fmt.Sprintf(
		"integration[%s]: %s",
		f.HandlerName,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}

// messageLoggedByProjection returns the log message for f.
func (l *Logger) messageLoggedByProjection(f MessageLoggedByProjection) string {
	return fmt.Sprintf(
		"projection[%s]: %s",
		f.HandlerName,
		fmt.Sprintf(f.LogFormat, f.LogArguments...),
	)
}
