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
	"github.com/onsi/gomega"
)

var _ = Context("message type assertions that match messages dispatched directly", func() {
	var (
		app            dogma.Application
		process        *ProcessMessageHandler
		command, event dogma.Message
		options        []engine.OperationOption
	)

	BeforeEach(func() {
		app, _, process, _ = newTestApp()
		command = MessageA{}
		event = MessageB{}
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
				t.Expect(
					testkit.Call(func() {
						if command != nil {
							err := t.CommandExecutor().ExecuteCommand(
								context.Background(),
								command,
							)
							gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
						}

						if event != nil {
							err := t.EventRecorder().RecordEvent(
								context.Background(),
								event,
							)
							gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
						}
					}),
					assertion,
				)
			},
			options,
			expectOk,
			expectReport,
		)
	}

	DescribeTable(
		"func CommandTypeExecuted()",
		test,
		Entry(
			"command type executed as expected",
			nil, // setup
			CommandTypeExecuted(MessageA{}),
			true, // ok
			`--- TEST REPORT ---`,
			``,
			`✓ execute any 'fixtures.MessageA' command`,
		),
		Entry(
			"no matching command types executed",
			nil, // setup
			CommandTypeExecuted(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any 'fixtures.MessageX' command`,
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
			CommandTypeExecuted(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any 'fixtures.MessageX' command`,
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
			CommandTypeExecuted(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     no commands were executed at all`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<process>' process message handler`,
			`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
		),
		Entry(
			"no matching command type executed and all relevant handler types disabled",
			func() {
				options = append(
					options,
					engine.EnableProcesses(false),
				)
			},
			CommandTypeExecuted(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any 'fixtures.MessageX' command`,
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
			// uniformity with the test suite for EventTypeRecorded(). Additionally,
			// the assertion report content will likely diverge from the test
			// above upon completion of https://github.com/dogmatiq/testkit/issues/66.
			"no matching command type executed and no relevant handler types engaged",
			func() {
				options = append(
					options,
					engine.EnableProcesses(false),
				)
			},
			CommandTypeExecuted(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     nothing executed the expected command`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable process handlers using the EnableHandlerType() option`,
			`  |     • verify the logic within the code that uses the dogma.CommandExecutor`,
		),
		Entry(
			"command of a similar type executed",
			nil,                              // setup
			CommandTypeExecuted(&MessageA{}), // note: message type is pointer
			false,                            // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any '*fixtures.MessageA' command`,
			``,
			`  | EXPLANATION`,
			`  |     a command of a similar type was executed via a dogma.CommandExecutor`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE TYPE DIFF`,
			`  |     [-*-]fixtures.MessageA`,
		),
		Entry(
			"expected message type recorded as an event rather than executed as a command",
			func() {
				command = nil
			},
			CommandTypeExecuted(MessageB{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any 'fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a message of this type was recorded as an event via a dogma.EventRecorder`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that an event of this type was intended to be recorded via a dogma.EventRecorder`,
			`  |     • verify that CommandTypeExecuted() is the correct assertion, did you mean EventTypeRecorded()?`,
		),
		Entry(
			"a message with a similar type recorded as an event rather than executed as a command",
			func() {
				command = nil
			},
			CommandTypeExecuted(&MessageB{}), // note: message type is pointer
			false,                            // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any '*fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a message of a similar type was recorded as an event via a dogma.EventRecorder`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that an event of this type was intended to be recorded via a dogma.EventRecorder`,
			`  |     • verify that CommandTypeExecuted() is the correct assertion, did you mean EventTypeRecorded()?`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE TYPE DIFF`,
			`  |     [-*-]fixtures.MessageB`,
		),
	)
})
