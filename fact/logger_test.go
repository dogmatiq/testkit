package fact_test

import (
	"errors"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/config"
	"github.com/dogmatiq/enginekit/config/runtimeconfig"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/testkit/envelope"
	. "github.com/dogmatiq/testkit/fact"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestLogger(t *testing.T) {
	t.Run("func Notify()", func(t *testing.T) {
		now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
		if err != nil {
			panic(err)
		}

		aggregate := runtimeconfig.FromAggregate(&AggregateMessageHandlerStub{
			ConfigureFunc: func(c dogma.AggregateConfigurer) {
				c.Identity("<aggregate>", "986495b4-c878-4e3a-b16b-73f8aefbc2ca")
				c.Routes(
					dogma.HandlesCommand[*CommandStub[TypeA]](),
					dogma.RecordsEvent[*EventStub[TypeA]](),
				)
			},
		})

		integration := runtimeconfig.FromIntegration(&IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<integration>", "2425a151-ba72-42ec-970a-8b3b4133b22f")
				c.Routes(
					dogma.HandlesCommand[*CommandStub[TypeA]](),
					dogma.RecordsEvent[*EventStub[TypeA]](),
				)
			},
		})

		process := runtimeconfig.FromProcess(&ProcessMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProcessConfigurer) {
				c.Identity("<process>", "570684db-0144-4628-a58f-ae815c55dea3")
				c.Routes(
					dogma.HandlesEvent[*EventStub[TypeA]](),
					dogma.ExecutesCommand[*CommandStub[TypeA]](),
				)
			},
		})

		projection := runtimeconfig.FromProjection(&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<projection>", "36f29880-6b87-42c5-848c-f515c9f1c74b")
				c.Routes(
					dogma.HandlesEvent[*EventStub[TypeA]](),
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
				HandlerType: config.ProcessHandlerType,
				InstanceID:  "<instance>",
			},
		)

		type logCase struct {
			Name    string
			Message string
			Fact    Fact
		}

		run := func(t *testing.T, cases []logCase) {
			for _, c := range cases {
				t.Run(c.Name, func(t *testing.T) {
					var (
						output string
						called bool
					)

					obs := NewLogger(func(s string) {
						called = true

						if output != "" {
							output += "\n"
						}

						output += s
					})

					obs.Notify(c.Fact)

					xtesting.Expect(t, "unexpected log output", output, c.Message)
					xtesting.Expect(t, "unexpected logger invocation", called, c.Message != "")
				})
			}
		}

		t.Run("dispatch", func(t *testing.T) {
			run(t, []logCase{
				{
					Name:    "DispatchCycleBegun",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ⚙    dispatching ● 2006-01-02T15:04:05+07:00 ● enabled: +aggregates +processes -<disabled> +<enabled>",
					Fact: DispatchCycleBegun{
						Envelope:   command,
						EngineTime: now,
						EnabledHandlerTypes: map[config.HandlerType]bool{
							config.AggregateHandlerType: true,
							config.ProcessHandlerType:   true,
						},
						EnabledHandlers: map[string]bool{
							"<enabled>":  true,
							"<disabled>": false,
						},
					},
				},
				{
					Name:    "DispatchCycleCompleted (success)",
					Message: "",
					Fact:    DispatchCycleCompleted{Envelope: command},
				},
				{
					Name:    "DispatchCycleCompleted (failure)",
					Message: "",
					Fact:    DispatchCycleCompleted{Envelope: command, Error: errors.New("<error>")},
				},
				{
					Name:    "DispatchBegun",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ⚙    *stubs.CommandStub[TypeA]? ● command(stubs.TypeA:A1, valid)",
					Fact:    DispatchBegun{Envelope: command},
				},
				{
					Name:    "CommandDeduplicated",
					Message: `= 10  ∵ 10  ⋲ 10  ↻ ⚙    *stubs.CommandStub[TypeA]? ● command ignored because it's idempotency key "<key>" has already been used`,
					Fact:    CommandDeduplicated{Envelope: command, Key: "<key>"},
				},
				{
					Name:    "DispatchCompleted (success)",
					Message: "",
					Fact:    DispatchCompleted{Envelope: command},
				},
				{
					Name:    "DispatchCompleted (failure)",
					Message: "",
					Fact:    DispatchCompleted{Envelope: command, Error: errors.New("<error>")},
				},
				{
					Name:    "HandlingBegun",
					Message: "",
					Fact:    HandlingBegun{},
				},
				{
					Name:    "HandlingCompleted (success)",
					Message: "",
					Fact:    HandlingCompleted{},
				},
				{
					Name:    "HandlingCompleted (failure)",
					Message: "= 10  ∵ 10  ⋲ 10  ▽ ∴ ✖  <aggregate> ● <error>",
					Fact:    HandlingCompleted{Handler: aggregate, Envelope: command, Error: errors.New("<error>")},
				},
				{
					Name:    "HandlingSkipped (handler type)",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because aggregate handlers are disabled",
					Fact:    HandlingSkipped{Handler: aggregate, Envelope: command, Reason: HandlerTypeDisabled},
				},
				{
					Name:    "HandlingSkipped (individual handler)",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because it is disabled during this tick of the test engine",
					Fact:    HandlingSkipped{Handler: aggregate, Envelope: command, Reason: IndividualHandlerDisabled},
				},
				{
					Name:    "HandlingSkipped (individual handler via configuration)",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> ● handler skipped because it is disabled by its Configure() method",
					Fact:    HandlingSkipped{Handler: aggregate, Envelope: command, Reason: IndividualHandlerDisabledByConfiguration},
				},
			})
		})

		t.Run("tick", func(t *testing.T) {
			run(t, []logCase{
				{
					Name:    "TickCycleBegun",
					Message: "= --  ∵ --  ⋲ --    ⚙    ticking ● 2006-01-02T15:04:05+07:00 ● enabled: +aggregates +processes -<disabled> +<enabled>",
					Fact: TickCycleBegun{
						EngineTime: now,
						EnabledHandlerTypes: map[config.HandlerType]bool{
							config.AggregateHandlerType: true,
							config.ProcessHandlerType:   true,
						},
						EnabledHandlers: map[string]bool{
							"<enabled>":  true,
							"<disabled>": false,
						},
					},
				},
				{
					Name:    "TickCycleCompleted (success)",
					Message: "",
					Fact:    TickCycleCompleted{},
				},
				{
					Name:    "TickCycleCompleted (failure)",
					Message: "",
					Fact:    TickCycleCompleted{Error: errors.New("<error>")},
				},
				{
					Name:    "TickBegun",
					Message: "",
					Fact:    TickBegun{},
				},
				{
					Name:    "TickCompleted (success)",
					Message: "",
					Fact:    TickCompleted{},
				},
				{
					Name:    "TickCompleted (failure)",
					Message: "= --  ∵ --  ⋲ --    ∴ ✖  <aggregate> ● <error>",
					Fact:    TickCompleted{Handler: aggregate, Error: errors.New("<error>")},
				},
			})
		})

		t.Run("aggregate", func(t *testing.T) {
			run(t, []logCase{
				{
					Name:    "AggregateInstanceLoaded",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● loaded an existing instance",
					Fact:    AggregateInstanceLoaded{Handler: aggregate, InstanceID: "<instance>", Envelope: command},
				},
				{
					Name:    "AggregateInstanceNotFound",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance does not yet exist",
					Fact:    AggregateInstanceNotFound{Handler: aggregate, InstanceID: "<instance>", Envelope: command},
				},
				{
					Name:    "AggregateInstanceCreated",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● instance created",
					Fact:    AggregateInstanceCreated{Handler: aggregate, InstanceID: "<instance>", Envelope: command},
				},
				{
					Name:    "EventRecordedByAggregate",
					Message: "= 20  ∵ 10  ⋲ 10  ▲ ∴    <aggregate> <instance> ● recorded an event ● *stubs.EventStub[TypeA]! ● event(stubs.TypeA:A1, valid)",
					Fact: EventRecordedByAggregate{
						Handler:    aggregate,
						InstanceID: "<instance>",
						Envelope:   command,
						EventEnvelope: command.NewEvent(
							"20",
							EventA1,
							time.Now(),
							envelope.Origin{},
							"a4dea2c6-6499-441c-94ad-686334880c1c",
							42,
						),
					},
				},
				{
					Name:    "MessageLoggedByAggregate",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ∴    <aggregate> <instance> ● <message>",
					Fact: MessageLoggedByAggregate{
						Handler:      aggregate,
						InstanceID:   "<instance>",
						Envelope:     command,
						LogFormat:    "<%s>",
						LogArguments: []any{"message"},
					},
				},
			})
		})

		t.Run("process", func(t *testing.T) {
			run(t, []logCase{
				{
					Name:    "ProcessInstanceLoaded",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● loaded an existing instance",
					Fact:    ProcessInstanceLoaded{Handler: process, InstanceID: "<instance>", Envelope: event},
				},
				{
					Name:    "ProcessEventIgnored",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> ● event ignored because it was not routed to any instance",
					Fact:    ProcessEventIgnored{Handler: process, Envelope: event},
				},
				{
					Name:    "ProcessEventRoutedToEndedInstance",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● event ignored because the target instance has ended",
					Fact:    ProcessEventRoutedToEndedInstance{Handler: process, InstanceID: "<instance>", Envelope: event},
				},
				{
					Name:    "ProcessTimeoutRoutedToEndedInstance",
					Message: "= 20  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● timeout ignored because the target instance has ended",
					Fact:    ProcessTimeoutRoutedToEndedInstance{Handler: process, InstanceID: "<instance>", Envelope: timeout},
				},
				{
					Name:    "ProcessInstanceNotFound",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance does not yet exist",
					Fact:    ProcessInstanceNotFound{Handler: process, InstanceID: "<instance>", Envelope: event},
				},
				{
					Name:    "ProcessInstanceBegun",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance begun",
					Fact:    ProcessInstanceBegun{Handler: process, InstanceID: "<instance>", Envelope: event},
				},
				{
					Name:    "ProcessInstanceEnded",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● instance ended",
					Fact:    ProcessInstanceEnded{Handler: process, InstanceID: "<instance>", Envelope: event},
				},
				{
					Name:    "ProcessInstanceEndingReverted",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● reverted ending process instance",
					Fact:    ProcessInstanceEndingReverted{Handler: process, InstanceID: "<instance>", Envelope: event},
				},
				{
					Name:    "CommandExecutedByProcess",
					Message: "= 20  ∵ 10  ⋲ 10  ▲ ≡    <process> <instance> ● executed a command ● *stubs.CommandStub[TypeA]? ● command(stubs.TypeA:A1, valid)",
					Fact: CommandExecutedByProcess{
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
				},
				{
					Name:    "TimeoutScheduledByProcess",
					Message: "= 20  ∵ 10  ⋲ 10  ▲ ≡    <process> <instance> ● scheduled a timeout for 2006-01-02T15:04:05+07:00 ● *stubs.TimeoutStub[TypeA]@ ● timeout(stubs.TypeA:A1, valid)",
					Fact:    TimeoutScheduledByProcess{Handler: process, InstanceID: "<instance>", TimeoutEnvelope: timeout},
				},
				{
					Name:    "MessageLoggedByProcess",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ≡    <process> <instance> ● <message>",
					Fact: MessageLoggedByProcess{
						Handler:      process,
						InstanceID:   "<instance>",
						Envelope:     event,
						LogFormat:    "<%s>",
						LogArguments: []any{"message"},
					},
				},
			})
		})

		t.Run("integration", func(t *testing.T) {
			run(t, []logCase{
				{
					Name:    "EventRecordedByIntegration",
					Message: "= 20  ∵ 10  ⋲ 10  ▲ ⨝    <integration> ● recorded an event ● *stubs.EventStub[TypeA]! ● event(stubs.TypeA:A1, valid)",
					Fact: EventRecordedByIntegration{
						Handler:  integration,
						Envelope: command,
						EventEnvelope: command.NewEvent(
							"20",
							EventA1,
							time.Now(),
							envelope.Origin{},
							"1494ce69-b98c-41b4-9617-7fa45aa1ed21",
							42,
						),
					},
				},
				{
					Name:    "MessageLoggedByIntegration",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ ⨝    <integration> ● <message>",
					Fact: MessageLoggedByIntegration{
						Handler:      integration,
						Envelope:     command,
						LogFormat:    "<%s>",
						LogArguments: []any{"message"},
					},
				},
			})
		})

		t.Run("projection", func(t *testing.T) {
			run(t, []logCase{
				{
					Name:    "ProjectionCompactionBegun",
					Message: "",
					Fact:    ProjectionCompactionBegun{},
				},
				{
					Name:    "ProjectionCompactionCompleted (success)",
					Message: "= --  ∵ --  ⋲ --    Σ    <projection> ● compacted",
					Fact:    ProjectionCompactionCompleted{Handler: projection},
				},
				{
					Name:    "ProjectionCompactionCompleted (failure)",
					Message: "= --  ∵ --  ⋲ --    Σ ✖  <projection> ● compaction failed: <error>",
					Fact:    ProjectionCompactionCompleted{Handler: projection, Error: errors.New("<error>")},
				},
				{
					Name:    "MessageLoggedByProjection",
					Message: "= 10  ∵ 10  ⋲ 10  ▼ Σ    <projection> ● <message>",
					Fact: MessageLoggedByProjection{
						Handler:      projection,
						Envelope:     command,
						LogFormat:    "<%s>",
						LogArguments: []any{"message"},
					},
				},
				{
					Name:    "MessageLoggedByProjection (compacting)",
					Message: "= --  ∵ --  ⋲ --    Σ    <projection> ● <message>",
					Fact: MessageLoggedByProjection{
						Handler:      projection,
						LogFormat:    "<%s>",
						LogArguments: []any{"message"},
					},
				},
			})
		})
	})
}
