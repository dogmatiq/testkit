package fact_test

import (
	"errors"
	"time"

	"github.com/dogmatiq/enginekit/fixtures"
	"github.com/dogmatiq/enginekit/handler"
	"github.com/dogmatiq/testkit/engine/envelope"
	. "github.com/dogmatiq/testkit/engine/fact"
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

		command := envelope.NewCommand(
			"100",
			fixtures.MessageC1,
			time.Now(),
		)

		event := envelope.NewEvent(
			"100",
			fixtures.MessageE1,
			time.Now(),
		)

		DescribeTable(
			"logs the expected message",
			func(m string, f Fact) {
				var (
					output string
					called bool
				)

				obs := NewLogger(
					func(s string) {
						called = true

						if output != "" {
							output += "\n"
						}

						output += s
					},
				)

				obs.Notify(f)

				Expect(output).To(BeIdenticalTo(m))
				Expect(called).To(Equal(m != ""))
			},

			// dispatch ...

			Entry(
				"DispatchCycleBegun",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ⚙    dispatching ● engine time is 2006-01-02T15:04:05+07:00 ● enabled: aggregate, process",
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
				"",
				DispatchCycleCompleted{
					Envelope: command,
				},
			),
			Entry(
				"DispatchCycleCompleted (failure)",
				"",
				DispatchCycleCompleted{
					Envelope: command,
					Error:    errors.New("<error>"),
				},
			),

			Entry(
				"DispatchBegun",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ⚙    fixtures.MessageC? ● {C1}",
				DispatchBegun{
					Envelope: command,
				},
			),
			Entry(
				"DispatchCompleted (success)",
				"",
				DispatchCompleted{
					Envelope: command,
				},
			),
			Entry(
				"DispatchCompleted (failure)",
				"",
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
				"= 0100  ∵ 0100  ⋲ 0100  ▽ ∴ ✖  <handler> ● <error>",
				HandlingCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Envelope:    command,
					Error:       errors.New("<error>"),
				},
			),
			Entry(
				"HandlingSkipped",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ∴    <handler> ● handler skipped because aggregate handlers are disabled",
				HandlingSkipped{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Envelope:    command,
				},
			),

			// tick ...

			Entry(
				"TickCycleBegun",
				"= ----  ∵ ----  ⋲ ----    ⚙    ticking ● engine time is 2006-01-02T15:04:05+07:00 ● enabled: aggregate, process",
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
				"",
				TickCycleCompleted{},
			),
			Entry(
				"TickCycleCompleted (failure)",
				"",
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
				"= ----  ∵ ----  ⋲ ----    ∴ ✖  <handler> ● <error>",
				TickCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Error:       errors.New("<error>"),
				},
			),

			// aggregates ...

			Entry(
				"AggregateInstanceLoaded",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ∴    <handler> <instance> ● loaded an existing instance",
				AggregateInstanceLoaded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"AggregateInstanceNotFound",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ∴    <handler> <instance> ● instance does not yet exist",
				AggregateInstanceNotFound{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"AggregateInstanceCreated",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ∴    <handler> <instance> ● instance created",
				AggregateInstanceCreated{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"AggregateInstanceDestroyed",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ∴    <handler> <instance> ● instance destroyed",
				AggregateInstanceDestroyed{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
				},
			),
			Entry(
				"EventRecordedByAggregate",
				"= 0200  ∵ 0100  ⋲ 0100  ▲ ∴    <handler> <instance> ● recorded an event ● fixtures.MessageE! ● {E1}",
				EventRecordedByAggregate{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    command,
					EventEnvelope: command.NewEvent(
						"200",
						fixtures.MessageE1,
						time.Now(),
						envelope.Origin{},
					),
				},
			),
			Entry(
				"MessageLoggedByAggregate",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ∴    <handler> <instance> ● <message>",
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
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ≡    <handler> <instance> ● loaded an existing instance",
				ProcessInstanceLoaded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessEventIgnored",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ≡    <handler> ● event ignored because it was not routed to any instance",
				ProcessEventIgnored{
					HandlerName: "<handler>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessTimeoutIgnored",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ≡    <handler> <instance> ● timeout ignored because the target instance no longer exists",
				ProcessTimeoutIgnored{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessInstanceNotFound",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ≡    <handler> <instance> ● instance does not yet exist",
				ProcessInstanceNotFound{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessInstanceBegun",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ≡    <handler> <instance> ● instance begun",
				ProcessInstanceBegun{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"ProcessInstanceEnded",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ≡    <handler> <instance> ● instance ended",
				ProcessInstanceEnded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
				},
			),
			Entry(
				"CommandExecutedByProcess",
				"= 0200  ∵ 0100  ⋲ 0100  ▲ ≡    <handler> <instance> ● executed a command ● fixtures.MessageC? ● {C1}",
				CommandExecutedByProcess{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					Envelope:    event,
					CommandEnvelope: event.NewCommand(
						"200",
						fixtures.MessageC1,
						time.Now(),
						envelope.Origin{},
					),
				},
			),
			Entry(
				"TimeoutScheduledByProcess",
				"= 0200  ∵ 0100  ⋲ 0100  ▲ ≡    <handler> <instance> ● scheduled a timeout for 2006-01-02T15:04:05+07:00 ● fixtures.MessageT@ ● {T1}",
				TimeoutScheduledByProcess{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					TimeoutEnvelope: event.NewTimeout(
						"200",
						fixtures.MessageT1,
						time.Now(),
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
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ≡    <handler> <instance> ● <message>",
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
				"= 0200  ∵ 0100  ⋲ 0100  ▲ ⨝    <handler> ● recorded an event ● fixtures.MessageE! ● {E1}",
				EventRecordedByIntegration{
					HandlerName: "<handler>",
					Envelope:    command,
					EventEnvelope: command.NewEvent(
						"200",
						fixtures.MessageE1,
						time.Now(),
						envelope.Origin{},
					),
				},
			),
			Entry(
				"MessageLoggedByIntegration",
				"= 0100  ∵ 0100  ⋲ 0100  ▼ ⨝    <handler> ● <message>",
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
				"= 0100  ∵ 0100  ⋲ 0100  ▼ Σ    <handler> ● <message>",
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
