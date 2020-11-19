package assert_test

import (
	"errors"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
)

var _ = Context("message assertions", func() {
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

	Describe("func CommandExecuted()", func() {
		It("panics if the message is invalid", func() {
			gomega.Expect(func() {
				CommandExecuted(MessageA{
					Value: errors.New("<invalid>"),
				})
			}).To(gomega.PanicWith("can not assert that this command will be executed, it is invalid: <invalid>"))
		})
	})

	Describe("func RecordEvent()", func() {
		It("panics if the message is invalid", func() {
			gomega.Expect(func() {
				EventRecorded(MessageA{
					Value: errors.New("<invalid>"),
				})
			}).To(gomega.PanicWith("can not assert that this event will be recorded, it is invalid: <invalid>"))
		})
	})

	DescribeTable(
		"func CommandExecuted()",
		test,
		Entry(
			"command executed as expected",
			nil, // setup
			CommandExecuted(MessageC{Value: "<value>"}),
			true, // ok
			`--- TEST REPORT ---`,
			``,
			`✓ execute a specific 'fixtures.MessageC' command`,
		),
		Entry(
			"no matching command executed",
			nil, // setup
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
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
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
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
			CommandExecuted(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     no commands were executed at all`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<process>' process message handler`,
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
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
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
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageX' command`,
			``,
			`  | EXPLANATION`,
			`  |     no relevant handler types were enabled`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable process handlers using the EnableHandlerType() option`,
		),
		Entry(
			"similar command executed with a different type",
			nil, // setup
			CommandExecuted(&MessageC{Value: "<value>"}), // note: message type is pointer
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific '*fixtures.MessageC' command`,
			``,
			`  | EXPLANATION`,
			`  |     a command of a similar type was executed by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     [-*-]fixtures.MessageC{`,
			`  |         Value: "<value>"`,
			`  |     }`,
		),
		Entry(
			"similar command executed with a different value",
			nil, // setup
			CommandExecuted(MessageC{Value: "<different>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageC' command`,
			``,
			`  | EXPLANATION`,
			`  |     a similar command was executed by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the content of the message`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     fixtures.MessageC{`,
			`  |         Value: "<[-differ-]{+valu+}e[-nt-]>"`,
			`  |     }`,
		),
		Entry(
			"expected message recorded as an event rather than executed as a command",
			nil, // setup
			CommandExecuted(MessageB{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     the expected message was recorded as an event by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<aggregate>' aggregate message handler intended to record an event of this type`,
			`  |     • verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?`,
		),
		Entry(
			"similar message with a different value recorded as an event rather than executed as a command",
			nil, // setup
			CommandExecuted(MessageB{Value: "<different>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific 'fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a similar message was recorded as an event by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<aggregate>' aggregate message handler intended to record an event of this type`,
			`  |     • verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     fixtures.MessageB{`,
			`  |         Value: "<[-differ-]{+valu+}e[-nt-]>"`,
			`  |     }`,
		),
		Entry(
			"similar message with a different type recorded as an event rather than executed as a command",
			nil, // setup
			CommandExecuted(&MessageB{Value: "<value>"}), // note: message type is pointer
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ execute a specific '*fixtures.MessageB' command`,
			``,
			`  | EXPLANATION`,
			`  |     a message of a similar type was recorded as an event by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<aggregate>' aggregate message handler intended to record an event of this type`,
			`  |     • verify that CommandExecuted() is the correct assertion, did you mean EventRecorded()?`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     [-*-]fixtures.MessageB{`,
			`  |         Value: "<value>"`,
			`  |     }`,
		),
	)

	DescribeTable(
		"func EventRecorded()",
		test,
		Entry(
			"event recorded as expected",
			nil, // setup
			EventRecorded(MessageB{Value: "<value>"}),
			true, // ok
			`--- TEST REPORT ---`,
			``,
			`✓ record a specific 'fixtures.MessageB' event`,
		),
		Entry(
			"no matching event recorded",
			nil, // setup
			EventRecorded(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     none of the engaged handlers recorded the expected event`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<aggregate>' aggregate message handler`,
			`  |     • verify the logic within the '<integration>' integration message handler`,
		),
		Entry(
			"no matching event recorded and all relevant handler types disabled",
			func() {
				options = append(
					options,
					engine.EnableAggregates(false),
					engine.EnableIntegrations(false),
				)
			},
			EventRecorded(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     no relevant handler types were enabled`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • enable aggregate handlers using the EnableHandlerType() option`,
			`  |     • enable integration handlers using the EnableHandlerType() option`,
		),
		Entry(
			"no matching event recorded and no relevant handler types engaged",
			func() {
				options = append(
					options,
					engine.EnableAggregates(false),
				)
			},
			EventRecorded(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageX' event`,
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
			EventRecorded(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageX' event`,
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
			EventRecorded(MessageX{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageX' event`,
			``,
			`  | EXPLANATION`,
			`  |     no events were recorded at all`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify the logic within the '<integration>' integration message handler`,
		),
		Entry(
			"similar event recorded with a different type",
			nil, // setup
			EventRecorded(&MessageB{Value: "<value>"}), // note: message type is pointer
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific '*fixtures.MessageB' event`,
			``,
			`  | EXPLANATION`,
			`  |     an event of a similar type was recorded by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     [-*-]fixtures.MessageB{`,
			`  |         Value: "<value>"`,
			`  |     }`,
		),
		Entry(
			"similar event recorded with a different value",
			nil, // setup
			EventRecorded(MessageB{Value: "<different>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageB' event`,
			``,
			`  | EXPLANATION`,
			`  |     a similar event was recorded by the '<aggregate>' aggregate message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • check the content of the message`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     fixtures.MessageB{`,
			`  |         Value: "<[-differ-]{+valu+}e[-nt-]>"`,
			`  |     }`,
		),
		Entry(
			"expected message executed as a command rather than recorded as an event",
			nil, // setup
			EventRecorded(MessageC{Value: "<value>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageC' event`,
			``,
			`  | EXPLANATION`,
			`  |     the expected message was executed as a command by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
			`  |     • verify that EventRecorded() is the correct assertion, did you mean CommandExecuted()?`,
		),
		Entry(
			"similar message with a different value executed as a command rather than recorded as an event",
			nil, // setup
			EventRecorded(MessageC{Value: "<different>"}),
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific 'fixtures.MessageC' event`,
			``,
			`  | EXPLANATION`,
			`  |     a similar message was executed as a command by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
			`  |     • verify that EventRecorded() is the correct assertion, did you mean CommandExecuted()?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     fixtures.MessageC{`,
			`  |         Value: "<[-differ-]{+valu+}e[-nt-]>"`,
			`  |     }`,
		),
		Entry(
			"similar message with a different type executed as a command rather than recorded as an event",
			nil, // setup
			EventRecorded(&MessageC{Value: "<value>"}), // note: message type is pointer
			false, // ok
			`--- TEST REPORT ---`,
			``,
			`✗ record a specific '*fixtures.MessageC' event`,
			``,
			`  | EXPLANATION`,
			`  |     a message of a similar type was executed as a command by the '<process>' process message handler`,
			`  | `,
			`  | SUGGESTIONS`,
			`  |     • verify that the '<process>' process message handler intended to execute a command of this type`,
			`  |     • verify that EventRecorded() is the correct assertion, did you mean CommandExecuted()?`,
			`  |     • check the message type, should it be a pointer?`,
			`  | `,
			`  | MESSAGE DIFF`,
			`  |     [-*-]fixtures.MessageC{`,
			`  |         Value: "<value>"`,
			`  |     }`,
		),
	)
})
