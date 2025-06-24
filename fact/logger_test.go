package fact_test

import (
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/testkit/envelope"
	. "github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("type Logger", func() {
	g.Describe("func Notify()", func() {
		now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
		if err != nil {
			panic(err)
		}

		aggregate := configkit.FromAggregate(&AggregateMessageHandlerStub{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "986495b4-c878-4e3a-b16b-73f8aefbc2ca")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
		})

		integration := configkit.FromIntegration(&IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "2425a151-ba72-42ec-970a-8b3b4133b22f")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
		})

		process := configkit.FromProcess(&ProcessMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "570684db-0144-4628-a58f-ae815c55dea3")
				c.Routes(
					dogma.HandlesEvent[EventStub[TypeA]](),
					dogma.ExecutesCommand[CommandStub[TypeA]](),
				)
			},
		})

		projection := configkit.FromProjection(&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "36f29880-6b87-42c5-848c-f515c9f1c74b")
				c.Routes(
					dogma.HandlesEvent[EventStub[TypeA]](),
				)
			},
		})

		command := envelope.NewCommand(
			"10",
			CommandA1,
			time.Now(),
		)

		event := envelope.NewEvent(
			"10",
			EventA1,
			time.Now(),
		)

		timeout := event.NewTimeout(
			"20",
			TimeoutA1,
			time.Now(),
			now,
			envelope.Origin{
				Handler:     process,
				HandlerType: configkit.ProcessHandlerType,
				InstanceID:  "<instance>",
			},
		)

		g.DescribeTable(
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

				gm.Expect(output).To(gm.BeIdenticalTo(m))
				gm.Expect(called).To(gm.Equal(m != ""))
			},

			// dispatch ...

			g.Entry(
				"DispatchCycleBegun",
				"= 10  ∵ 10  ⋲ 10  ▼ ⚙    dispatching ● 2006-01-02T15:04:05+07:00 ● enabled: +aggregates +processes -<disabled> +<enabled>",
				DispatchCycleBegun{
					Envelope:   command,
					EngineTime: now,
					EnabledHandlerTypes: map[configkit.HandlerType]bool{
						configkit.AggregateHandlerType: true,
						configkit.ProcessHandlerType:   true,
					},
					EnabledHandlers: map[string]bool{
						"<enabled>":  true,
						"<disabled>": false,
					},
				},
			),
			g.Entry(
				"DispatchCycleCompleted (success)",
				"",
				DispatchCycleCompleted{
					Envelope: command,
				},
			),
			g.Entry(
				"DispatchCycleCompleted (failure)",
				"",
				DispatchCycleCompleted{
					Envelope: command,
					Error:    errors.New("<error>"),
				},
			),

			g.Entry(
				"DispatchBegun",
				"= 10  ∵ 10  ⋲ 10  ▼ ⚙    stubs.CommandStub[TypeA]? ● command(stubs.TypeA:A1, valid)",
				DispatchBegun{
					Envelope: command,
				},
			),
			g.Entry(
				"DispatchCompleted (success)",
				"",
				DispatchCompleted{
					Envelope: command,
				},
			),
			g.Entry(
				"DispatchCompleted (failure)",
				"",
				DispatchCompleted{
					Envelope: command,
					Error:    errors.New("<error>"),
				},
			),

			g.Entry(
				"HandlingBegun",
				"",
				HandlingBegun{},
			),
			g.Entry(
				"HandlingCompleted (success)",
				"",
				HandlingCompleted{},
			),
			g.Entry(
				"HandlingCompleted (failure)",
				"= 10  ∵ 10  ⋲ 10  ▽ ∴ ✖  <aggregate> ● <error>",
				HandlingCompleted{
					Handler:  aggregate,
					Envelope: command,
					Error:    errors.New("<error>"),
				},
			),
			g.Entry(
				"HandlingSkipped (handler type)",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because aggregate handlers are disabled",
				HandlingSkipped{
					Handler:  aggregate,
					Envelope: command,
					Reason:   HandlerTypeDisabled,
				},
			),
			g.Entry(
				"HandlingSkipped (individual handler)",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because it is disabled during this tick of the test engine",
				HandlingSkipped{
					Handler:  aggregate,
					Envelope: command,
					Reason:   IndividualHandlerDisabled,
				},
			),
			g.Entry(
				"HandlingSkipped (individual handler via configuration)",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because it is disabled by its Configure() method",
				HandlingSkipped{
					Handler:  aggregate,
					Envelope: command,
					Reason:   IndividualHandlerDisabledByConfiguration,
				},
			),

			// tick ...

			g.Entry(
				"TickCycleBegun",
				"= --  ∵ --  ⋲ --    ⚙    ticking ● 2006-01-02T15:04:05+07:00 ● enabled: +aggregates +processes -<disabled> +<enabled>",
				TickCycleBegun{
					EngineTime: now,
					EnabledHandlerTypes: map[configkit.HandlerType]bool{
						configkit.AggregateHandlerType: true,
						configkit.ProcessHandlerType:   true,
					},
					EnabledHandlers: map[string]bool{
						"<enabled>":  true,
						"<disabled>": false,
					},
				},
			),
			g.Entry(
				"TickCycleCompleted (success)",
				"",
				TickCycleCompleted{},
			),
			g.Entry(
				"TickCycleCompleted (failure)",
				"",
				TickCycleCompleted{
					Error: errors.New("<error>"),
				},
			),

			g.Entry(
				"TickBegun",
				"",
				TickBegun{},
			),
			g.Entry(
				"TickCompleted (success)",
				"",
				TickCompleted{},
			),
			g.Entry(
				"TickCompleted (failure)",
				"= --  ∵ --  ⋲ --    ∴ ✖  <aggregate> ● <error>",
				TickCompleted{
					Handler: aggregate,
					Error:   errors.New("<error>"),
				},
			),

			// aggregates ...

			g.Entry(
				"AggregateInstanceLoaded",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● loaded an existing instance",
				AggregateInstanceLoaded{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			g.Entry(
				"AggregateInstanceNotFound",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance does not yet exist",
				AggregateInstanceNotFound{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			g.Entry(
				"AggregateInstanceCreated",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance created",
				AggregateInstanceCreated{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			g.Entry(
				"AggregateInstanceDestroyed",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance destroyed",
				AggregateInstanceDestroyed{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			g.Entry(
				"AggregateInstanceDestructionReverted",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● destruction of instance reverted",
				AggregateInstanceDestructionReverted{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			g.Entry(
				"EventRecordedByAggregate",
				"= 20  ∵ 10  ⋲ 10  ▲ ∴    <aggregate> <instance> ● recorded an event ● stubs.EventStub[TypeA]! ● event(stubs.TypeA:A1, valid)",
				EventRecordedByAggregate{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
					EventEnvelope: command.NewEvent(
						"20",
						EventA1,
						time.Now(),
						envelope.Origin{},
					),
				},
			),
			g.Entry(
				"MessageLoggedByAggregate",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● <message>",
				MessageLoggedByAggregate{
					Handler:      aggregate,
					InstanceID:   "<instance>",
					Envelope:     command,
					LogFormat:    "<%s>",
					LogArguments: []any{"message"},
				},
			),

			// processes ...

			g.Entry(
				"ProcessInstanceLoaded",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● loaded an existing instance",
				ProcessInstanceLoaded{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			g.Entry(
				"ProcessEventIgnored",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> ● event ignored because it was not routed to any instance",
				ProcessEventIgnored{
					Handler:  process,
					Envelope: event,
				},
			),
			g.Entry(
				"ProcessEventRoutedToEndedInstance",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● event ignored because the target instance has ended",
				ProcessEventRoutedToEndedInstance{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			g.Entry(
				"ProcessTimeoutRoutedToEndedInstance",
				"= 20  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● timeout ignored because the target instance has ended",
				ProcessTimeoutRoutedToEndedInstance{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   timeout,
				},
			),
			g.Entry(
				"ProcessInstanceNotFound",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance does not yet exist",
				ProcessInstanceNotFound{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			g.Entry(
				"ProcessInstanceBegun",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance begun",
				ProcessInstanceBegun{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			g.Entry(
				"ProcessInstanceEnded",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance ended",
				ProcessInstanceEnded{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			g.Entry(
				"ProcessInstanceEndingReverted",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● reverted ending process instance",
				ProcessInstanceEndingReverted{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			g.Entry(
				"CommandExecutedByProcess",
				"= 20  ∵ 10  ⋲ 10  ▲ ≡    <process> <instance> ● executed a command ● stubs.CommandStub[TypeA]? ● command(stubs.TypeA:A1, valid)",
				CommandExecutedByProcess{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
					CommandEnvelope: event.NewCommand(
						"20",
						CommandA1,
						time.Now(),
						envelope.Origin{},
					),
				},
			),
			g.Entry(
				"TimeoutScheduledByProcess",
				"= 20  ∵ 10  ⋲ 10  ▲ ≡    <process> <instance> ● scheduled a timeout for 2006-01-02T15:04:05+07:00 ● stubs.TimeoutStub[TypeA]@ ● timeout(stubs.TypeA:A1, valid)",
				TimeoutScheduledByProcess{
					Handler:         process,
					InstanceID:      "<instance>",
					TimeoutEnvelope: timeout,
				},
			),
			g.Entry(
				"MessageLoggedByProcess",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● <message>",
				MessageLoggedByProcess{
					Handler:      process,
					InstanceID:   "<instance>",
					Envelope:     event,
					LogFormat:    "<%s>",
					LogArguments: []any{"message"},
				},
			),

			// integrations ...

			g.Entry(
				"EventRecordedByIntegration",
				"= 20  ∵ 10  ⋲ 10  ▲ ⨝    <integration> ● recorded an event ● stubs.EventStub[TypeA]! ● event(stubs.TypeA:A1, valid)",
				EventRecordedByIntegration{
					Handler:  integration,
					Envelope: command,
					EventEnvelope: command.NewEvent(
						"20",
						EventA1,
						time.Now(),
						envelope.Origin{},
					),
				},
			),
			g.Entry(
				"MessageLoggedByIntegration",
				"= 10  ∵ 10  ⋲ 10  ▼ ⨝    <integration> ● <message>",
				MessageLoggedByIntegration{
					Handler:      integration,
					Envelope:     command,
					LogFormat:    "<%s>",
					LogArguments: []any{"message"},
				},
			),

			// projections ...

			g.Entry(
				"ProjectionCompactionBegun",
				"",
				ProjectionCompactionBegun{},
			),

			g.Entry(
				"ProjectionCompactionCompleted (success)",
				"= --  ∵ --  ⋲ --    Σ    <projection> ● compacted",
				ProjectionCompactionCompleted{
					Handler: projection,
				},
			),

			g.Entry(
				"ProjectionCompactionCompleted (failure)",
				"= --  ∵ --  ⋲ --    Σ ✖  <projection> ● compaction failed: <error>",
				ProjectionCompactionCompleted{
					Handler: projection,
					Error:   errors.New("<error>"),
				},
			),

			g.Entry(
				"MessageLoggedByProjection",
				"= 10  ∵ 10  ⋲ 10  ▼ Σ    <projection> ● <message>",
				MessageLoggedByProjection{
					Handler:      projection,
					Envelope:     command,
					LogFormat:    "<%s>",
					LogArguments: []any{"message"},
				},
			),

			g.Entry(
				"MessageLoggedByProjection (compacting)",
				"= --  ∵ --  ⋲ --    Σ    <projection> ● <message>",
				MessageLoggedByProjection{
					Handler:      projection,
					LogFormat:    "<%s>",
					LogArguments: []any{"message"},
				},
			),
		)
	})
})
