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

		DescribeTable(
			"logs the expected message",
			func(m string, f Fact) {
				var output string

				obs := &Logger{
					Log: func(s string, v ...interface{}) {
						output = fmt.Sprintf(s, v...)
					},
				}

				obs.Notify(f)

				Expect(output).To(Equal(m))
			},

			// dispatch ...

			Entry(
				"DispatchCycleBegun",
				"engine: dispatch of 'fixtures.MessageA' command begun at 2006-01-02T15:04:05+07:00 (enabled: aggregate, process)",
				DispatchCycleBegun{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
					EngineTime: now,
					EnabledHandlers: map[handler.Type]bool{
						handler.AggregateType: true,
						handler.ProcessType:   true,
					},
				},
			),
			Entry(
				"DispatchCycleCompleted (success)",
				"engine: dispatch of 'fixtures.MessageA' command completed successfully",
				DispatchCycleCompleted{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
				},
			),
			Entry(
				"DispatchCycleCompleted (failure)",
				"engine: dispatch of 'fixtures.MessageA' command completed with errors",
				DispatchCycleCompleted{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
					Error: errors.New("<error>"),
				},
			),
			Entry(
				"DispatchCycleSkipped",
				"engine: no route for 'fixtures.MessageA' messages",
				DispatchCycleSkipped{
					Message: fixtures.MessageA1,
				},
			),

			XEntry(
				"DispatchBegun",
				"engine: dispatch of 'fixtures.MessageA' command begun at 2006-01-02T15:04:05+07:00 (enabled: aggregate, process)",
				DispatchBegun{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
				},
			),
			XEntry(
				"DispatchCompleted (success)",
				"engine: dispatch of 'fixtures.MessageA' command completed successfully",
				DispatchCompleted{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
				},
			),
			XEntry(
				"DispatchCompleted (failure)",
				"engine: dispatch of 'fixtures.MessageA' command completed with errors",
				DispatchCompleted{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
					Error: errors.New("<error>"),
				},
			),

			Entry(
				"HandlingBegun",
				"aggregate[<handler>]: message handling begun",
				HandlingBegun{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"HandlingCompleted (success)",
				"aggregate[<handler>]: handled message successfully",
				HandlingCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"HandlingCompleted (failure)",
				"aggregate[<handler>]: handling failed: <error>",
				HandlingCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Error:       errors.New("<error>"),
				},
			),
			Entry(
				"HandlingSkipped",
				"aggregate[<handler>]: message handling skipped because aggregate handlers are disabled",
				HandlingSkipped{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),

			// tick ...

			Entry(
				"TickCycleBegun",
				"engine: tick begun at 2006-01-02T15:04:05+07:00 (enabled: aggregate, process)",
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
				"engine: tick completed successfully",
				TickCycleCompleted{},
			),
			Entry(
				"TickCycleCompleted (failure)",
				"engine: tick completed with errors",
				TickCycleCompleted{
					Error: errors.New("<error>"),
				},
			),

			Entry(
				"TickBegun",
				"aggregate[<handler>]: tick begun",
				TickBegun{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"TickCompleted (success)",
				"aggregate[<handler>]: tick completed successfully",
				TickCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"TickCompleted (failure)",
				"aggregate[<handler>]: tick failed: <error>",
				TickCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Error:       errors.New("<error>"),
				},
			),

			// aggregates ...

			Entry(
				"AggregateInstanceLoaded",
				"aggregate[<handler>@<instance>]: loading existing instance",
				AggregateInstanceLoaded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"AggregateInstanceNotFound",
				"aggregate[<handler>@<instance>]: no existing instance found",
				AggregateInstanceNotFound{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"AggregateInstanceCreated",
				"aggregate[<handler>@<instance>]: instance created",
				AggregateInstanceCreated{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"AggregateInstanceDestroyed",
				"aggregate[<handler>@<instance>]: instance destroyed",
				AggregateInstanceDestroyed{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"EventRecordedByAggregate",
				"aggregate[<handler>@<instance>]: recorded 'fixtures.MessageA' event",
				EventRecordedByAggregate{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					EventEnvelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.EventRole,
					),
				},
			),
			Entry(
				"MessageLoggedByAggregate",
				"aggregate[<handler>@<instance>]: <message>",
				MessageLoggedByAggregate{
					HandlerName:  "<handler>",
					InstanceID:   "<instance>",
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),

			// processes ...

			Entry(
				"ProcessInstanceLoaded",
				"process[<handler>@<instance>]: loading existing instance",
				ProcessInstanceLoaded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"ProcessEventIgnored",
				"process[<handler>]: event not routed to any instance",
				ProcessEventIgnored{
					HandlerName: "<handler>",
				},
			),
			Entry(
				"ProcessTimeoutIgnored",
				"process[<handler>@<instance>]: timeout's instance no longer exists",
				ProcessTimeoutIgnored{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"ProcessInstanceNotFound",
				"process[<handler>@<instance>]: no existing instance found",
				ProcessInstanceNotFound{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"ProcessInstanceBegun",
				"process[<handler>@<instance>]: instance begun",
				ProcessInstanceBegun{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"ProcessInstanceEnded",
				"process[<handler>@<instance>]: instance ended",
				ProcessInstanceEnded{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
				},
			),
			Entry(
				"CommandExecutedByProcess",
				"process[<handler>@<instance>]: executed 'fixtures.MessageA' command",
				CommandExecutedByProcess{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					CommandEnvelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
				},
			),
			Entry(
				"TimeoutScheduledByProcess",
				"process[<handler>@<instance>]: scheduled 'fixtures.MessageT' timeout for 2006-01-02T15:04:05+07:00",
				TimeoutScheduledByProcess{
					HandlerName: "<handler>",
					InstanceID:  "<instance>",
					TimeoutEnvelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.EventRole,
					).NewTimeout(
						2000,
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
				"process[<handler>@<instance>]: <message>",
				MessageLoggedByProcess{
					HandlerName:  "<handler>",
					InstanceID:   "<instance>",
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),

			// integrations ...

			Entry(
				"EventRecordedByIntegration",
				"integration[<handler>]: recorded 'fixtures.MessageA' event",
				EventRecordedByIntegration{
					HandlerName: "<handler>",
					EventEnvelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.EventRole,
					),
				},
			),
			Entry(
				"MessageLoggedByIntegration",
				"integration[<handler>]: <message>",
				MessageLoggedByIntegration{
					HandlerName:  "<handler>",
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),

			// projections ...

			Entry(
				"MessageLoggedByProjection",
				"projection[<handler>]: <message>",
				MessageLoggedByProjection{
					HandlerName:  "<handler>",
					LogFormat:    "<%s>",
					LogArguments: []interface{}{"message"},
				},
			),
		)
	})
})
