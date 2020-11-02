package assert_test

import (
	"context"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Context("message assertions that match messages dispatched directly", func() {
	var (
		app            dogma.Application
		process        *ProcessMessageHandler
		command, event dogma.Message
		options        []engine.OperationOption
	)

	BeforeEach(func() {
		app, _, process, _ = newTestApp()
		command = MessageA{Value: "<value>"}
		event = MessageB{Value: "<value>"}
		options = nil
	})

	test := func(
		setup func(),
		assertion Assertion,
		expectOk bool,
		expectReport ...string,
	) {
		if setup != nil {
			setup()
		}

		runTest(
			app,
			func(t *testkit.Test) {
				t.Call(
					func() error {
						if command != nil {
							return t.CommandExecutor().ExecuteCommand(
								context.Background(),
								command,
							)
						}

						if event != nil {
							return t.EventRecorder().RecordEvent(
								context.Background(),
								event,
							)
						}

						return nil
					},
					assertion,
				)
			},
			options,
			expectOk,
			expectReport,
		)
	}

	DescribeTable(
		"func CommandExecuted()",
		test,
		Entry(
			"command executed as expected",
			nil, // setup
			CommandExecuted(MessageC{Value: "<value>"}),
			true, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✓ execute a specific 'fixtures.MessageC' command`,
		),
		Entry(
			"no matching command executed",
			nil, // setup
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     nothing executed the expected command`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<process>' process message handler`,
			`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
		),
		Entry(
			"no messages produced at all",
			func() {
				process.HandleEventFunc = nil
				command = nil
				event = nil
			},
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     no messages were produced at all`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
		),
		Entry(
			"no commands produced at all",
			func() {
				process.HandleEventFunc = nil
				command = nil
			},
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     no commands were executed at all`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<process>' process message handler`,
			`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
		),
		Entry(
			"no matching command executed and all relevant handler types disabled",
			func() {
				options = append(
					options,
					engine.EnableProcesses(false),
				)
			},
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     nothing executed the expected command`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable process handlers using the EnableHandlerType() option`,
			`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
		),
		Entry(
			// Note: the report produced from this test is actually the same as
			// the test above because there is only one relevant handler type
			// (process) that can be disabled. It is kept for completeness and
			// uniformity with the test suite for EventRecorded(). Additionally,
			// the assertion report content will likely diverge from the test
			// above upon completion of https://github.com/dogmatiq/testkit/issues/66.
			"no matching command executed and no relevant handler types engaged",
			func() {
				options = append(
					options,
					engine.EnableProcesses(false),
				)
			},
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     nothing executed the expected command`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable process handlers using the EnableHandlerType() option`,
			`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
		),
		Entry(
			"similar command executed with a different type",
			nil, // setup
			CommandExecuted(&MessageA{Value: "<value>"}), // note: message type is pointer
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific '*fixtures.MessageA' command`,
			``,
			`  | EXPLANATION`,
			`  |     a command of a similar type was executed via a dogma.CommandExecutor`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     [-*-]fixtures.MessageA{`,
			`  |         Value: "<value>"`,
			`  |     }`,
		),
		Entry(
			"similar command executed with a different value",
			nil, // setup
			CommandExecuted(MessageA{Value: "<different>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageA' command`,
			``,
			`  | EXPLANATION`,
			`  |     a similar command was executed via a dogma.CommandExecutor`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the content of the message`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     fixtures.MessageA{`,
			`  |         Value: "<[-differ-]{+valu+}e[-nt-]>"`,
			`  |     }`,
		),
		Entry(
			"expected message recorded as an event rather than executed as a command",
			func() {
				command = nil
			},
			CommandExecuted(MessageB{Value: "<value>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     the expected message was recorded as an event via a dogma.EventRecorder`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that an event of this type was intended to be recorded via a dogma.EventRecorder`,
			`  |     • verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?`,
		),
		Entry(
			"similar message with a different value recorded as an event rather than executed as a command",
			func() {
				command = nil
			},
			CommandExecuted(MessageB{Value: "<different>"}),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a similar message was recorded as an event via a dogma.EventRecorder`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that an event of this type was intended to be recorded via a dogma.EventRecorder`,
			`  |     • verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     fixtures.MessageB{`,
			`  |         Value: "<[-differ-]{+valu+}e[-nt-]>"`,
			`  |     }`,
		),
		Entry(
			"similar message with a different type recorded as an event rather than executed as a command",
			func() {
				command = nil
			},
			CommandExecuted(&MessageB{Value: "<value>"}), // note: message type is pointer
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ execute a specific '*fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a message of a similar type was recorded as an event via a dogma.EventRecorder`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that an event of this type was intended to be recorded via a dogma.EventRecorder`,
			`  |     • verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     [-*-]fixtures.MessageB{`,
			`  |         Value: "<value>"`,
			`  |     }`,
		),
	)
})
