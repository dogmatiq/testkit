package assert_test

import (
	"errors"
	"strings"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/assert"
	. "github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
)

var _ = Describe("type userAssertion", func() {
	var app dogma.Application

	BeforeEach(func() {
		app = &Application{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "<app-key>")

				c.RegisterAggregate(&AggregateMessageHandler{
					ConfigureFunc: func(c dogma.AggregateConfigurer) {
						c.Identity("<aggregate>", "<aggregate-key>")
						c.ConsumesCommandType(MessageA{})
						c.ProducesEventType(MessageB{})
					},
					RouteCommandToInstanceFunc: func(dogma.Message) string {
						return "<aggregate-instance>"
					},
				})
			},
		}
	})

	test := func(
		assertion Assertion,
		ok bool,
		report ...string,
	) {
		t := &testingmock.T{
			FailSilently: true,
		}

		testkit.
			New(app).
			Begin(t).
			ExecuteCommand(
				MessageA{},
				assertion,
			)

		logs := strings.TrimSpace(strings.Join(t.Logs, "\n"))
		lines := strings.Split(logs, "\n")

		gomega.Expect(lines).To(gomega.Equal(report))
		gomega.Expect(t.Failed).To(gomega.Equal(!ok))
	}

	DescribeTable(
		"func Should()",
		test,
		Entry(
			"assertion passed",
			assert.Should(
				"<criteria>",
				func(AssertionContext) error {
					return nil
				},
			),
			true, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✓ <criteria>`,
		),
		Entry(
			"assertion failed",
			assert.Should(
				"<criteria>",
				func(AssertionContext) error {
					return errors.New("<explanation>")
				},
			),
			false, // ok
			`--- ASSERTION REPORT ---`,
			``,
			`✗ <criteria> (the user-defined assertion returned a non-nil error)`,
			``,
			`  | EXPLANATION`,
			`  |     <explanation>`,
		),
	)
})
