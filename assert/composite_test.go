package assert_test

import (
	"errors"
	"strings"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
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

	Describe("func AllOf()", func() {
		It("panics if no sub-assertions are provided", func() {
			gomega.Expect(func() {
				AllOf()
			}).To(gomega.Panic())
		})

		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"single sub-assertion is flattened",
				AllOf(
					Should(
						"<sub-criteria>",
						func(AssertionContext) error {
							return nil
						},
					),
				),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <sub-criteria>`,
			),
			Entry(
				"all sub-assertions passed",
				AllOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return nil
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return nil
						},
					),
				),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ all of`,
				`    ✓ <sub-criteria 1>`,
				`    ✓ <sub-criteria 2>`,
			),
			Entry(
				"some of the sub-assertions passed",
				AllOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return nil
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return errors.New("<explanation>")
						},
					),
				),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ all of (1 of the sub-assertions failed)`,
				`    ✓ <sub-criteria 1>`,
				`    ✗ <sub-criteria 2> (the user-defined assertion returned a non-nil error)`,
				`    `,
				`      | EXPLANATION`,
				`      |     <explanation>`,
			),
			Entry(
				"none of the sub-assertions passed",
				AllOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return errors.New("<explanation 1>")
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return errors.New("<explanation 2>")
						},
					),
				),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ all of (2 of the sub-assertions failed)`,
				`    ✗ <sub-criteria 1> (the user-defined assertion returned a non-nil error)`,
				`    `,
				`      | EXPLANATION`,
				`      |     <explanation 1>`,
				`    `,
				`    ✗ <sub-criteria 2> (the user-defined assertion returned a non-nil error)`,
				`    `,
				`      | EXPLANATION`,
				`      |     <explanation 2>`,
			),
		)
	})

	Describe("func AnyOf()", func() {
		It("panics if no sub-assertions are provided", func() {
			gomega.Expect(func() {
				AnyOf()
			}).To(gomega.Panic())
		})

		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"single sub-assertion is flattened",
				AnyOf(
					Should(
						"<sub-criteria>",
						func(AssertionContext) error {
							return nil
						},
					),
				),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <sub-criteria>`,
			),
			Entry(
				"all sub-assertions passed",
				AnyOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return nil
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return nil
						},
					),
				),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ any of`,
				`    ✓ <sub-criteria 1>`,
				`    ✓ <sub-criteria 2>`,
			),
			Entry(
				"some of the sub-assertions passed",
				AnyOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return nil
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return errors.New("<explanation>")
						},
					),
				),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ any of`,
				`    ✓ <sub-criteria 1>`,
				`    ✗ <sub-criteria 2>`,
			),
			Entry(
				"none of the sub-assertions passed",
				AnyOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return errors.New("<explanation 1>")
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return errors.New("<explanation 2>")
						},
					),
				),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ any of (all 2 of the sub-assertions failed)`,
				`    ✗ <sub-criteria 1> (the user-defined assertion returned a non-nil error)`,
				`    `,
				`      | EXPLANATION`,
				`      |     <explanation 1>`,
				`    `,
				`    ✗ <sub-criteria 2> (the user-defined assertion returned a non-nil error)`,
				`    `,
				`      | EXPLANATION`,
				`      |     <explanation 2>`,
			),
		)
	})

	Describe("func NoneOf()", func() {
		It("panics if no sub-assertions are provided", func() {
			gomega.Expect(func() {
				NoneOf()
			}).To(gomega.Panic())
		})

		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"single sub-assertion is not flattened",
				NoneOf(
					Should(
						"<sub-criteria>",
						func(AssertionContext) error {
							return nil
						},
					),
				),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ none of (the sub-assertion passed unexpectedly)`,
				`    ✓ <sub-criteria>`,
			),
			Entry(
				"all sub-assertions passed",
				NoneOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return nil
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return nil
						},
					),
				),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ none of (2 of the sub-assertions passed unexpectedly)`,
				`    ✓ <sub-criteria 1>`,
				`    ✓ <sub-criteria 2>`,
			),
			Entry(
				"some of the sub-assertions passed",
				NoneOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return nil
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return errors.New("<explanation>")
						},
					),
				),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ none of (1 of the sub-assertions passed unexpectedly)`,
				`    ✓ <sub-criteria 1>`,
				`    ✗ <sub-criteria 2> (the user-defined assertion returned a non-nil error)`,
				`    `,
				`      | EXPLANATION`,
				`      |     <explanation>`,
			),
			Entry(
				"none of the sub-assertions passed",
				NoneOf(
					Should(
						"<sub-criteria 1>",
						func(AssertionContext) error {
							return errors.New("<explanation 1>")
						},
					),
					Should(
						"<sub-criteria 2>",
						func(AssertionContext) error {
							return errors.New("<explanation 2>")
						},
					),
				),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ none of`,
				`    ✗ <sub-criteria 1>`,
				`    ✗ <sub-criteria 2>`,
			),
		)
	})
})
