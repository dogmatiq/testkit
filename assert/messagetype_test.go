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
		integration *IntegrationMessageHandler
		action      testkit.Action
		options     []engine.OperationOption
	)

	BeforeEach(func() {
		app, aggregate, _, integration = newTestApp()
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
