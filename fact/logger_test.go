package fact_test

import (
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit/envelope"
	. "github.com/dogmatiq/testkit/fact"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("type Logger", func() {
	g.Describe("func Notify()", func() {
		now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
		if err != nil {
			panic(err)
		}

		command := envelope.NewCommand(
			"10",
			MessageC1,
			time.Now(),
		)

		event := envelope.NewEvent(
			"10",
			MessageE1,
			time.Now(),
		)

		aggregate := configkit.FromAggregate(&AggregateMessageHandler{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "986495b4-c878-4e3a-b16b-73f8aefbc2ca")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
		})

		integration := configkit.FromIntegration(&IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "2425a151-ba72-42ec-970a-8b3b4133b22f")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
		})

		process := configkit.FromProcess(&ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "570684db-0144-4628-a58f-ae815c55dea3")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
			},
		})

		projection := configkit.FromProjection(&ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "36f29880-6b87-42c5-848c-f515c9f1c74b")
				c.ConsumesEventType(MessageE{})
			},
		})

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

				Expect(output).To(BeIdenticalTo(m))
				Expect(called).To(Equal(m != ""))
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
				"= 10  ∵ 10  ⋲ 10  ▼ ⚙    fixtures.MessageC? ● {C1}",
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
				"HandlingSkipped",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because aggregate handlers are disabled",
				HandlingSkipped{
					Handler:  aggregate,
					Envelope: command,
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
				"= 20  ∵ 10  ⋲ 10  ▲ ∴    <aggregate> <instance> ● recorded an event ● fixtures.MessageE! ● {E1}",
				EventRecordedByAggregate{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
					EventEnvelope: command.NewEvent(
						"20",
						MessageE1,
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
					LogArguments: []interface{}{"message"},
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
				"ProcessTimeoutIgnored",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● timeout ignored because the target instance no longer exists",
				ProcessTimeoutIgnored{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
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
				"= 20  ∵ 10  ⋲ 10  ▲ ≡    <process> <instance> ● executed a command ● fixtures.MessageC? ● {C1}",
				CommandExecutedByProcess{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
					CommandEnvelope: event.NewCommand(
						"20",
						MessageC1,
						time.Now(),
						envelope.Origin{},
					),
				},
			),
			g.Entry(
				"TimeoutScheduledByProcess",
				"= 20  ∵ 10  ⋲ 10  ▲ ≡    <process> <instance> ● scheduled a timeout for 2006-01-02T15:04:05+07:00 ● fixtures.MessageT@ ● {T1}",
				TimeoutScheduledByProcess{
					Handler:    process,
					InstanceID: "<instance>",
					TimeoutEnvelope: event.NewTimeout(
						"20",
						MessageT1,
						time.Now(),
						now,
						envelope.Origin{
							Handler:     process,
							HandlerType: configkit.ProcessHandlerType,
							InstanceID:  "<instance>",
						},
					),
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
					LogArguments: []interface{}{"message"},
				},
			),

			// integrations ...

			g.Entry(
				"EventRecordedByIntegration",
				"= 20  ∵ 10  ⋲ 10  ▲ ⨝    <integration> ● recorded an event ● fixtures.MessageE! ● {E1}",
				EventRecordedByIntegration{
					Handler:  integration,
					Envelope: command,
					EventEnvelope: command.NewEvent(
						"20",
						MessageE1,
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
					LogArguments: []interface{}{"message"},
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
					LogArguments: []interface{}{"message"},
				},
			),

			g.Entry(
				"MessageLoggedByProjection (compacting)",
				"= --  ∵ --  ⋲ --    Σ    <projection> ● <message>",
				MessageLoggedByProjection{
					Handler:      projection,
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),
		)
	})
})
