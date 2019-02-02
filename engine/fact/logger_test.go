package fact_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/dogmatiq/dogmatest/engine/envelope"
	. "github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/enginekit/fixtures"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Logger", func() {
	Describe("func Notify()", func() {
		now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
		if err != nil {
			panic(err)
		}

		command := envelope.New(
			"1000",
			fixtures.MessageC1,
			message.CommandRole,
		)

		event := envelope.New(
			"1000",
			fixtures.MessageE1,
			message.EventRole,
		)

		DescribeTable(
			"logs the expected message",
			func(m string, f Fact) {
				var (
					output string
					called bool
				)

				obs := &Logger{
					Log: func(s string, v ...interface{}) {
						called = true
						output = fmt.Sprintf(s, v...)
					},
				}

				obs.Notify(f)

				Expect(output).To(BeIdenticalTo(m))
				Expect(called).To(Equal(m != ""))
			},

			// dispatch ...

			Entry(
				"DispatchCycleBegun",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ⚙    dispatch cycle begun at 2006-01-02T15:04:05+07:00 [enabled: aggregate, process]",
				DispatchCycleBegun{
					Envelope:   command,
					EngineTime: now,
					EnabledHandlers: map[handler.Type]bool{
						handler.AggregateType: true,
						handler.ProcessType:   true,
					},
				},
			),
			Entry(
				"DispatchCycleCompleted (success)",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ⚙    dispatch cycle completed successfully",
				DispatchCycleCompleted{
					Envelope: command,
				},
			),
			Entry(
				"DispatchCycleCompleted (failure)",
				"= 1000  ∵ 1000  ⋲ 1000  ▽ ⚙ ✖  dispatch cycle completed with errors",
				DispatchCycleCompleted{
					Envelope: command,
					Error:    errors.New("<error>"),
				},
			),
			Entry(
				"DispatchCycleSkipped",
				"= ----  ∵ ----  ⋲ ----  ▼ ⚙    fixtures.MessageC ● dispatch cycle skipped because this message type is not routed to any handlers",
				DispatchCycleSkipped{
					Message: fixtures.MessageC1,
				},
			),

			Entry(
				"DispatchBegun",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ⚙    fixtures.MessageC? ● {C1} ● dispatch begun",
				DispatchBegun{
					Envelope: command,
				},
			),
			Entry(
				"DispatchCompleted (success)",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ⚙    dispatch completed successfully",
				DispatchCompleted{
					Envelope: command,
				},
			),
			Entry(
				"DispatchCompleted (failure)",
				"= 1000  ∵ 1000  ⋲ 1000  ▽ ⚙ ✖  dispatch completed with errors",
				DispatchCompleted{
					Envelope: command,
					Error:    errors.New("<error>"),
				},
			),

			Entry(
				"HandlingBegun",
				"",
				HandlingBegun{},
			),
			Entry(
				"HandlingCompleted (success)",
				"",
				HandlingCompleted{},
			),
			Entry(
				"HandlingCompleted (failure)",
				"= 1000  ∵ 1000  ⋲ 1000  ▽ ∴ ✖  [<handler>]  <error>",
				HandlingCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Envelope:    command,
					Error:       errors.New("<error>"),
				},
			),
			Entry(
				"HandlingSkipped",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ∴    [<handler>]  handler skipped because aggregate handlers are disabled",
				HandlingSkipped{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Envelope:    command,
				},
			),

			// tick ...

			Entry(
				"TickCycleBegun",
				"= ----  ∵ ----  ⋲ ----    ⚙    tick cycle begun at 2006-01-02T15:04:05+07:00 [enabled: aggregate, process]",
				TickCycleBegun{
					EngineTime: now,
					EnabledHandlers: map[handler.Type]bool{
						handler.AggregateType: true,
						handler.ProcessType:   true,
					},
				},
			),
			Entry(
				"TickCycleCompleted (success)",
				"= ----  ∵ ----  ⋲ ----    ⚙    tick cycle completed successfully",
				TickCycleCompleted{},
			),
			Entry(
				"TickCycleCompleted (failure)",
				"= ----  ∵ ----  ⋲ ----    ⚙ ✖  tick cycle completed with errors",
				TickCycleCompleted{
					Error: errors.New("<error>"),
				},
			),

			Entry(
				"TickBegun",
				"",
				TickBegun{},
			),
			Entry(
				"TickCompleted (success)",
				"",
				TickCompleted{},
			),
			Entry(
				"TickCompleted (failure)",
				"= ----  ∵ ----  ⋲ ----    ∴ ✖  [<handler>]  <error>",
				TickCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Error:       errors.New("<error>"),
				},
			),

			// aggregates ...

			Entry(
				"AggregateInstanceLoaded",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ∴    [<handler> <instance>]  loaded an existing instance",
				AggregateInstanceLoaded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"AggregateInstanceNotFound",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ∴    [<handler> <instance>]  instance does not yet exist",
				AggregateInstanceNotFound{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"AggregateInstanceCreated",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ∴    [<handler> <instance>]  instance created",
				AggregateInstanceCreated{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"AggregateInstanceDestroyed",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ∴    [<handler> <instance>]  instance destroyed",
				AggregateInstanceDestroyed{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"EventRecordedByAggregate",
				"= 2000  ∵ 1000  ⋲ 1000  ▲ ∴    [<handler> <instance>]  recorded an event ● fixtures.MessageE! ● {E1}",
				EventRecordedByAggregate{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
					EventEnvelope: command.NewEvent(
						"2000",
						fixtures.MessageE1,
						envelope.Origin{},
					),
				},
			),
			Entry(
				"MessageLoggedByAggregate",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ∴    [<handler> <instance>]  <message>",
				MessageLoggedByAggregate{
					HandlerName:  "<handler>",
					InstanceID:   "<instance>",
					Envelope:     command,
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),

			// processes ...

			Entry(
				"ProcessInstanceLoaded",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ☰    [<handler> <instance>]  loaded an existing instance",
				ProcessInstanceLoaded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessEventIgnored",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ☰    [<handler>]  event ignored because it was not routed to any instance",
				ProcessEventIgnored{
					HandlerName: "<handler>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessTimeoutIgnored",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ☰    [<handler> <instance>]  timeout ignored because the target instance no longer exists",
				ProcessTimeoutIgnored{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessInstanceNotFound",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ☰    [<handler> <instance>]  instance does not yet exist",
				ProcessInstanceNotFound{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessInstanceBegun",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ☰    [<handler> <instance>]  instance begun",
				ProcessInstanceBegun{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessInstanceEnded",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ☰    [<handler> <instance>]  instance ended",
				ProcessInstanceEnded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"CommandExecutedByProcess",
				"= 2000  ∵ 1000  ⋲ 1000  ▲ ☰    [<handler> <instance>]  executed a command ● fixtures.MessageC? ● {C1}",
				CommandExecutedByProcess{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
					CommandEnvelope: event.NewCommand(
						"2000",
						fixtures.MessageC1,
						envelope.Origin{},
					),
				},
			),
			Entry(
				"TimeoutScheduledByProcess",
				"= 2000  ∵ 1000  ⋲ 1000  ▲ ☰    [<handler> <instance>]  scheduled a timeout for 2006-01-02T15:04:05+07:00 ● fixtures.MessageT@ ● {T1}",
				TimeoutScheduledByProcess{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					TimeoutEnvelope: event.NewTimeout(
						"2000",
						fixtures.MessageT1,
						now,
						envelope.Origin{
							HandlerName: "<handler>",
							HandlerType: handler.ProcessType,
							InstanceID:  "<instance>",
						},
					),
				},
			),
			Entry(
				"MessageLoggedByProcess",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ☰    [<handler> <instance>]  <message>",
				MessageLoggedByProcess{
					HandlerName:  "<handler>",
					InstanceID:   "<instance>",
					Envelope:     event,
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),

			// integrations ...

			Entry(
				"EventRecordedByIntegration",
				"= 2000  ∵ 1000  ⋲ 1000  ▲ ⨝    [<handler>]  recorded an event ● fixtures.MessageE! ● {E1}",
				EventRecordedByIntegration{
					HandlerName: "<handler>",
					Envelope:    command,
					EventEnvelope: command.NewEvent(
						"2000",
						fixtures.MessageE1,
						envelope.Origin{},
					),
				},
			),
			Entry(
				"MessageLoggedByIntegration",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ ⨝    [<handler>]  <message>",
				MessageLoggedByIntegration{
					HandlerName:  "<handler>",
					Envelope:     command,
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),

			// projections ...

			Entry(
				"MessageLoggedByProjection",
				"= 1000  ∵ 1000  ⋲ 1000  ▼ Σ    [<handler>]  <message>",
				MessageLoggedByProjection{
					HandlerName:  "<handler>",
					Envelope:     command,
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),
		)
	})
})
