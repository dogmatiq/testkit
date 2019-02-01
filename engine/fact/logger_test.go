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
				"UnroutableMessageDispatched",
				"engine: no route for 'fixtures.MessageA' messages",
				UnroutableMessageDispatched{
					Message: fixtures.MessageA1,
				},
			),
			Entry(
				"MessageDispatchBegun",
				"engine: dispatch of 'fixtures.MessageA' command begun at 2006-01-02T15:04:05+07:00 (enabled: aggregate, process)",
				MessageDispatchBegun{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
					Now: now,
					EnabledHandlers: map[handler.Type]bool{
						handler.AggregateType: true,
						handler.ProcessType:   true,
					},
				},
			),
			Entry(
				"MessageDispatchCompleted (success)",
				"engine: dispatch of 'fixtures.MessageA' command completed successfully",
				MessageDispatchCompleted{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
				},
			),
			Entry(
				"MessageDispatchCompleted (failure)",
				"engine: dispatch of 'fixtures.MessageA' command completed with errors",
				MessageDispatchCompleted{
					Envelope: envelope.New(
						1000,
						fixtures.MessageA1,
						message.CommandRole,
					),
					Error: errors.New("<error>"),
				},
			),

			Entry(
				"MessageHandlingBegun",
				"aggregate[<handler>]: message handling begun",
				MessageHandlingBegun{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"MessageHandlingCompleted (success)",
				"aggregate[<handler>]: handled message successfully",
				MessageHandlingCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"MessageHandlingCompleted (failure)",
				"aggregate[<handler>]: handling failed: <error>",
				MessageHandlingCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
					Error:       errors.New("<error>"),
				},
			),
			Entry(
				"MessageHandlingSkipped",
				"aggregate[<handler>]: message handling skipped because aggregate handlers are disabled",
				MessageHandlingSkipped{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),

			// tick ...

			Entry(
				"EngineTickBegun",
				"engine: tick begun at 2006-01-02T15:04:05+07:00 (enabled: aggregate, process)",
				EngineTickBegun{
					Now: now,
					EnabledHandlers: map[handler.Type]bool{
						handler.AggregateType: true,
						handler.ProcessType:   true,
					},
				},
			),
			Entry(
				"EngineTickCompleted (success)",
				"engine: tick completed successfully",
				EngineTickCompleted{},
			),
			Entry(
				"EngineTickCompleted (failure)",
				"engine: tick completed with errors",
				EngineTickCompleted{
					Error: errors.New("<error>"),
				},
			),

			Entry(
				"ControllerTickBegun",
				"aggregate[<handler>]: tick begun",
				ControllerTickBegun{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"ControllerTickCompleted (success)",
				"aggregate[<handler>]: tick completed successfully",
				ControllerTickCompleted{
					HandlerName: "<handler>",
					HandlerType: handler.AggregateType,
				},
			),
			Entry(
				"ControllerTickCompleted (failure)",
				"aggregate[<handler>]: tick failed: <error>",
				ControllerTickCompleted{
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
