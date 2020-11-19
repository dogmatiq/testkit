package assert_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Context("message type assertions", func() {
	var (
		app         dogma.Application
		aggregate   *AggregateMessageHandler
		process     *ProcessMessageHandler
		integration *IntegrationMessageHandler
		action      testkit.Action
		options     []engine.OperationOption
	)

	BeforeEach(func() {
		app, aggregate, process, integration = newTestApp()
		action = testkit.ExecuteCommand(MessageA{})
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
				t.Expect(action, assertion)
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
			CommandTypeExecuted(MessageC{}),
			true, // ok
			`--- TEST REPORT ---`,
			``,
			`✓ execute any 'fixtures.MessageC' command`,
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
			`  |     none of the engaged handlers executed the expected command`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<process>' process message handler`,
		),
		Entry(
			"no messages produced at all",
			func() {
				process.HandleEventFunc = nil
				action = testkit.RecordEvent(MessageB{})
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
			`  |     • verify the logic within the '<process>' process message handler`,
		),
		Entry(
			"no commands produced at all",
			func() {
				process.HandleEventFunc = nil
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
			`  |     no relevant handler types were enabled`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable process handlers using the EnableHandlerType() option`,
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
			`  |     no relevant handler types were enabled`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable process handlers using the EnableHandlerType() option`,
		),
		Entry(
			"command of a similar type executed",
			nil,                              // setup
			CommandTypeExecuted(&MessageC{}), // note: message type is pointer
			false,                            // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any '*fixtures.MessageC' command`,
			``,
			`  | EXPLANATION`,
			`  |     a command of a similar type was executed by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE TYPE DIFF`,
			`  |     [-*-]fixtures.MessageC`,
		),
		Entry(
			"expected message type recorded as an event rather than executed as a command",
			nil, // setup
			CommandTypeExecuted(MessageB{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any 'fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a message of this type was recorded as an event by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<aggregate>' aggregate message handler intended to record an event of this type`,
			`  |     • verify that CommandTypeExecuted() is the correct assertion, did you mean EventTypeRecorded()?`,
		),
		Entry(
			"a message with a similar type recorded as an event rather than executed as a command",
			nil,                              // setup
			CommandTypeExecuted(&MessageB{}), // note: message type is pointer
			false,                            // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute any '*fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a message of a similar type was recorded as an event by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<aggregate>' aggregate message handler intended to record an event of this type`,
			`  |     • verify that CommandTypeExecuted() is the correct assertion, did you mean EventTypeRecorded()?`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE TYPE DIFF`,
			`  |     [-*-]fixtures.MessageB`,
		),
	)

	DescribeTable(
		"func EventTypeRecorded()",
		test,
		Entry(
			"event type recorded as expected",
			nil, // setup
			EventTypeRecorded(MessageB{}),
			true, // ok
			`--- TEST REPORT ---`,
			``,
			`✓ record any 'fixtures.MessageB' event`,
		),
		Entry(
			"no matching event type recorded",
			nil, // setup
			EventTypeRecorded(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     none of the engaged handlers recorded the expected event`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			`  |     • verify the logic within the '<integration>' integration message handler`,
		),
		Entry(
			"no matching event type recorded and all relevant handler types disabled",
			func() {
				options = append(
					options,
					engine.EnableAggregates(false),
					engine.EnableIntegrations(false),
				)
			},
			EventTypeRecorded(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     no relevant handler types were enabled`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable aggregate handlers using the EnableHandlerType() option`,
			`  |     • enable integration handlers using the EnableHandlerType() option`,
		),
		Entry(
			"no matching event type recorded and no relevant handler types engaged",
			func() {
				options = append(
					options,
					engine.EnableAggregates(false),
				)
			},
			EventTypeRecorded(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     no relevant handlers (aggregate or integration) were engaged`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable aggregate handlers using the EnableHandlerType() option`,
			`  |     • check the application's routing configuration`,
		),
		Entry(
			"no messages produced at all",
			func() {
				aggregate.HandleCommandFunc = nil
			},
			EventTypeRecorded(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     no messages were produced at all`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
		),
		Entry(
			"no events recorded at all",
			func() {
				integration.HandleCommandFunc = nil
				action = testkit.RecordEvent(MessageB{})
			},
			EventTypeRecorded(MessageX{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     no events were recorded at all`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<integration>' integration message handler`,
		),
		Entry(
			"event of a similar type recorded",
			nil,                            // setup
			EventTypeRecorded(&MessageB{}), // note: message type is pointer
			false,                          // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any '*fixtures.MessageB' event`,
			``,
			`  | EXPLANATION`,
			`  |     an event of a similar type was recorded by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE TYPE DIFF`,
			`  |     [-*-]fixtures.MessageB`,
		),
		Entry(
			"expected message type executed as a command rather than recorded as an event",
			nil, // setup
			EventTypeRecorded(MessageC{}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any 'fixtures.MessageC' event`,
			``,
			`  | EXPLANATION`,
			`  |     a message of this type was executed as a command by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
			`  |     • verify that EventTypeRecorded() is the correct assertion, did you mean CommandTypeExecuted()?`,
		),
		Entry(
			"a message with a similar type executed as a command rather than recorded as an event",
			nil,                            // setup
			EventTypeRecorded(&MessageC{}), // note: message type is pointer
			false,                          // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record any '*fixtures.MessageC' event`,
			``,
			`  | EXPLANATION`,
			`  |     a message of a similar type was executed as a command by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
			`  |     • verify that EventTypeRecorded() is the correct assertion, did you mean CommandTypeExecuted()?`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE TYPE DIFF`,
			`  |     [-*-]fixtures.MessageC`,
		),
	)
})
