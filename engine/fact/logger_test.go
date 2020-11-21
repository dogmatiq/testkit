package fact_test

import (
	"errors"
	"time"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
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
				c.Identity("<aggregate>", "<aggregate-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
		})

		integration := configkit.FromIntegration(&IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "<integration-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
			},
		})

		process := configkit.FromProcess(&ProcessMessageHandler{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "<process-key>")
				c.ConsumesEventType(MessageE{})
				c.ProducesCommandType(MessageC{})
			},
		})

		projection := configkit.FromProjection(&ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "<projection-key>")
				c.ConsumesEventType(MessageE{})
			},
		})

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
				"= 10  ∵ 10  ⋲ 10  ▼ ⚙    fixtures.MessageC? ● {C1}",
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
				"= 10  ∵ 10  ⋲ 10  ▽ ∴ ✖  <aggregate> ● <error>",
				HandlingCompleted{
					Handler:  aggregate,
					Envelope: command,
					Error:    errors.New("<error>"),
				},
			),
			Entry(
				"HandlingSkipped",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because aggregate handlers are disabled",
				HandlingSkipped{
					Handler:  aggregate,
					Envelope: command,
				},
			),

			// tick ...

			Entry(
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
				"= --  ∵ --  ⋲ --    ∴ ✖  <aggregate> ● <error>",
				TickCompleted{
					Handler: aggregate,
					Error:   errors.New("<error>"),
				},
			),

			// aggregates ...

			Entry(
				"AggregateInstanceLoaded",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● loaded an existing instance",
				AggregateInstanceLoaded{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			Entry(
				"AggregateInstanceNotFound",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance does not yet exist",
				AggregateInstanceNotFound{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			Entry(
				"AggregateInstanceCreated",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance created",
				AggregateInstanceCreated{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			Entry(
				"AggregateInstanceDestroyed",
				"= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance destroyed",
				AggregateInstanceDestroyed{
					Handler:    aggregate,
					InstanceID: "<instance>",
					Envelope:   command,
				},
			),
			Entry(
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
			Entry(
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

			Entry(
				"ProcessInstanceLoaded",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● loaded an existing instance",
				ProcessInstanceLoaded{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			Entry(
				"ProcessEventIgnored",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> ● event ignored because it was not routed to any instance",
				ProcessEventIgnored{
					Handler:  process,
					Envelope: event,
				},
			),
			Entry(
				"ProcessTimeoutIgnored",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● timeout ignored because the target instance no longer exists",
				ProcessTimeoutIgnored{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			Entry(
				"ProcessInstanceNotFound",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance does not yet exist",
				ProcessInstanceNotFound{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			Entry(
				"ProcessInstanceBegun",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance begun",
				ProcessInstanceBegun{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			Entry(
				"ProcessInstanceEnded",
				"= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance ended",
				ProcessInstanceEnded{
					Handler:    process,
					InstanceID: "<instance>",
					Envelope:   event,
				},
			),
			Entry(
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
			Entry(
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
			Entry(
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

			Entry(
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
			Entry(
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

			Entry(
				"ProjectionCompactionBegun",
				"",
				ProjectionCompactionBegun{},
			),

			Entry(
				"ProjectionCompactionCompleted (success)",
				"= --  ∵ --  ⋲ --    Σ    <projection> ● compacted",
				ProjectionCompactionCompleted{
					Handler: projection,
				},
			),

			Entry(
				"ProjectionCompactionCompleted (failure)",
				"= --  ∵ --  ⋲ --    Σ ✖  <projection> ● compaction failed: <error>",
				ProjectionCompactionCompleted{
					Handler: projection,
					Error:   errors.New("<error>"),
				},
			),

			Entry(
				"MessageLoggedByProjection",
				"= 10  ∵ 10  ⋲ 10  ▼ Σ    <projection> ● <message>",
				MessageLoggedByProjection{
					Handler:      projection,
					Envelope:     command,
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),

			Entry(
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
