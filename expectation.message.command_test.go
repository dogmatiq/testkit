package testkit_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/engine"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestToExecuteCommand(t *testing.T) {
	type (
		EventThatIsIgnored        = EventStub[TypeX]
		EventThatExecutesCommand  = EventStub[TypeC]
		EventThatSchedulesTimeout = EventStub[TypeT]

		CommandThatIsExecuted      = CommandStub[TypeC]
		CommandThatIsNeverExecuted = CommandStub[TypeX]
		CommandThatIsOnlyConsumed  = CommandStub[TypeO]

		TimeoutThatIsScheduled = TimeoutStub[TypeT]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "ce773269-4ad7-4c7f-a0ff-cda2e5899743")

			c.Routes(
				dogma.ViaProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
						c.Routes(
							dogma.HandlesEvent[*EventThatIsIgnored](),

							dogma.HandlesEvent[*EventThatExecutesCommand](),
							dogma.ExecutesCommand[*CommandThatIsExecuted](),
							dogma.ExecutesCommand[*CommandThatIsNeverExecuted](),

							dogma.HandlesEvent[*EventThatSchedulesTimeout](),
							dogma.SchedulesTimeout[*TimeoutThatIsScheduled](),
						)
					},
					RouteEventToInstanceFunc: func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "<instance>", true, nil
					},
					HandleEventFunc: func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
						m dogma.Event,
					) error {
						switch m := m.(type) {
						case *EventThatExecutesCommand:
							s.ExecuteCommand(
								&CommandThatIsExecuted{
									Content: m.Content,
								},
							)
						case *EventThatSchedulesTimeout:
							s.ScheduleTimeout(
								&TimeoutThatIsScheduled{
									Content: m.Content,
								},
								time.Now().Add(1*time.Hour),
							)
						}

						return nil
					},
				}),

				dogma.ViaIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "49fa7c5f-7682-4743-bf8a-ed96dee2d81a")
						c.Routes(
							dogma.HandlesCommand[*CommandThatIsOnlyConsumed](),
						)
					},
				}),
			)
		},
	}

	cases := []struct {
		Name        string
		Action      func(*testing.T, *Test) Action
		Expectation Expectation
		Passes      bool
		Report      reportMatcher
		Options     []TestOption
	}{
		{
			"command executed as expected",
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommand(&CommandThatIsExecuted{}),
			true,
			expectReport(
				`✓ execute a specific '*stubs.CommandStub[TypeC]' command`,
			),
			nil,
		},
		{
			"no matching command executed",
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommand(&CommandThatIsNeverExecuted{}),
			false,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeX]' command`,
				``,
				`  | EXPLANATION`,
				`  |     none of the engaged handlers executed a matching command`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no messages produced at all",
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatIsIgnored{})
			},
			ToExecuteCommand(&CommandThatIsExecuted{}),
			false,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no messages were produced at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no commands produced at all",
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatSchedulesTimeout{})
			},
			ToExecuteCommand(&CommandThatIsExecuted{}),
			false,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no commands were executed at all`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • verify the logic within the '<process>' process message handler`,
			),
			nil,
		},
		{
			"no matching command executed and all relevant handler types disabled",
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			ToExecuteCommand(&CommandThatIsExecuted{}),
			false,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     no relevant handler types were enabled`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • enable process handlers using the EnableHandlerType() option`,
			),
			[]TestOption{
				WithUnsafeOperationOptions(
					engine.EnableProcesses(false),
				),
			},
		},
		{
			"similar command executed with a different value",
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{Content: "<content>"})
			},
			ToExecuteCommand(&CommandThatIsExecuted{Content: "<different>"}),
			false,
			expectReport(
				`✗ execute a specific '*stubs.CommandStub[TypeC]' command`,
				``,
				`  | EXPLANATION`,
				`  |     a similar command was executed by the '<process>' process message handler`,
				`  | `,
				`  | SUGGESTIONS`,
				`  |     • check the content of the message`,
				`  | `,
				`  | MESSAGE DIFF`,
				`  |     *stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeC]{`,
				`  |         Content:         "<[-differ-]{+cont+}ent>"`,
				`  |         ValidationError: ""`,
				`  |     }`,
			),
			nil,
		},
		{
			"does not include an explanation when negated and a sibling expectation passes",
			func(t *testing.T, tc *Test) Action {
				return RecordEvent(&EventThatExecutesCommand{})
			},
			NoneOf(
				ToExecuteCommand(&CommandThatIsExecuted{}),
				ToExecuteCommand(&CommandThatIsNeverExecuted{}),
			),
			false,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ execute a specific '*stubs.CommandStub[TypeC]' command`,
				`    ✗ execute a specific '*stubs.CommandStub[TypeX]' command`,
			),
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mt := &testingmock.T{FailSilently: true}
			tc := Begin(mt, app, c.Options...)
			tc.Expect(c.Action(t, tc), c.Expectation)
			c.Report(mt)
			if mt.Failed() != !c.Passes {
				t.Fatalf(
					"expectation should have %s but %s",
					map[bool]string{true: "passed", false: "failed"}[c.Passes],
					map[bool]string{true: "passed", false: "failed"}[mt.Failed()],
				)
			}
		})
	}
}

func TestToExecuteCommand_UnrecognizedMessageType(t *testing.T) {
	type (
		EventThatIsIgnored        = EventStub[TypeX]
		CommandThatIsOnlyConsumed = CommandStub[TypeO]
	)

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "ce773269-4ad7-4c7f-a0ff-cda2e5899743")
			c.Routes(
				dogma.ViaProcess(&ProcessMessageHandlerStub{
					ConfigureFunc: func(c dogma.ProcessConfigurer) {
						c.Identity("<process>", "8b4c4701-be92-4b28-83b6-0d69b97fb451")
						c.Routes(
							dogma.HandlesEvent[*EventThatIsIgnored](),
							dogma.ExecutesCommand[*CommandThatIsOnlyConsumed](),
						)
					},
					RouteEventToInstanceFunc: func(
						context.Context,
						dogma.Event,
					) (string, bool, error) {
						return "<instance>", true, nil
					},
					HandleEventFunc: func(
						_ context.Context,
						_ dogma.ProcessRoot,
						s dogma.ProcessEventScope,
						m dogma.Event,
					) error {
						return nil
					},
				}),
			)
		},
	}

	mt := &testingmock.T{FailSilently: true}
	Begin(mt, app).Expect(
		noop,
		ToExecuteCommand(CommandU1),
	)

	xtesting.Expect(t, "test should have failed", mt.Failed(), true)
	if !slices.Contains(mt.Logs, "  |     a command of type *stubs.CommandStub[TypeU] can never be executed, the application does not use this message type") {
		// REVIEW: why are these pipes now necessary in the assertion?
		t.Fatalf("expected unrecognized message type error in logs: %v", mt.Logs)
	}
}

func TestToExecuteCommand_UnproducedMessageType(t *testing.T) {
	type CommandThatIsOnlyConsumed = CommandStub[TypeO]

	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "ce773269-4ad7-4c7f-a0ff-cda2e5899743")
			c.Routes(
				dogma.ViaIntegration(&IntegrationMessageHandlerStub{
					ConfigureFunc: func(c dogma.IntegrationConfigurer) {
						c.Identity("<integration>", "49fa7c5f-7682-4743-bf8a-ed96dee2d81a")
						c.Routes(
							dogma.HandlesCommand[*CommandThatIsOnlyConsumed](),
						)
					},
				}),
			)
		},
	}

	mt := &testingmock.T{FailSilently: true}
	Begin(mt, app).Expect(
		noop,
		ToExecuteCommand(&CommandThatIsOnlyConsumed{}),
	)

	xtesting.Expect(t, "test should have failed", mt.Failed(), true)
	if !slices.Contains(mt.Logs, "  |     no handlers execute commands of type *stubs.CommandStub[TypeO], it is only ever consumed") {
		t.Fatalf("expected unproduced message type error in logs: %v", mt.Logs)
	}
}

func TestToExecuteCommand_NilMessage(t *testing.T) {
	xtesting.ExpectPanic(t, "ToExecuteCommand(<nil>): message must not be nil", func() {
		ToExecuteCommand(nil)
	})
}

func TestToExecuteCommand_InvalidMessage(t *testing.T) {
	xtesting.ExpectPanic(t, "ToExecuteCommand(*stubs.CommandStub[TypeA]): <invalid>", func() {
		ToExecuteCommand(&CommandStub[TypeA]{
			ValidationError: "<invalid>",
		})
	})
}
