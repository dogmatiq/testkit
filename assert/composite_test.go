package assert_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
)

var _ = Context("composite assertions", func() {
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
		expectOk bool,
		expectReport ...string,
	) {
		runTest(
			app,
			func(t *testkit.Test) {
				t.ExecuteCommand(MessageA{}, assertion)
			},
			nil, //options
			expectOk,
			expectReport,
		)
	}

	Describe("func AllOf()", func() {
		It("panics if no sub-assertions are provided", func() {
			gomega.Expect(func() {
				AllOf()
			}).To(gomega.PanicWith("no sub-assertions provided"))
		})

		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"single sub-assertion is flattened",
				AllOf(pass),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <always pass>`,
			),
			Entry(
				"all sub-assertions passed",
				AllOf(pass, pass),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ all of`,
				`    ✓ <always pass>`,
				`    ✓ <always pass>`,
			),
			Entry(
				"some of the sub-assertions passed",
				AllOf(pass, fail),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ all of (1 of the sub-assertions failed)`,
				`    ✓ <always pass>`,
				`    ✗ <always fail>`,
			),
			Entry(
				"none of the sub-assertions passed",
				AllOf(fail, fail),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ all of (2 of the sub-assertions failed)`,
				`    ✗ <always fail>`,
				`    ✗ <always fail>`,
			),
		)
	})

	Describe("func AnyOf()", func() {
		It("panics if no sub-assertions are provided", func() {
			gomega.Expect(func() {
				AnyOf()
			}).To(gomega.PanicWith("no sub-assertions provided"))
		})

		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"single sub-assertion is flattened",
				AnyOf(pass),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ <always pass>`,
			),
			Entry(
				"all sub-assertions passed",
				AnyOf(pass, pass),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ any of`,
				`    ✓ <always pass>`,
				`    ✓ <always pass>`,
			),
			Entry(
				"some of the sub-assertions passed",
				AnyOf(pass, fail),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ any of`,
				`    ✓ <always pass>`,
				`    ✗ <always fail>`,
			),
			Entry(
				"none of the sub-assertions passed",
				AnyOf(fail, fail),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ any of (all 2 of the sub-assertions failed)`,
				`    ✗ <always fail>`,
				`    ✗ <always fail>`,
			),
		)
	})

	Describe("func NoneOf()", func() {
		It("panics if no sub-assertions are provided", func() {
			gomega.Expect(func() {
				NoneOf()
			}).To(gomega.PanicWith("no sub-assertions provided"))
		})

		DescribeTable(
			"assertion reports",
			test,
			Entry(
				"single sub-assertion is not flattened",
				NoneOf(pass),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ none of (the sub-assertion passed unexpectedly)`,
				`    ✓ <always pass>`,
			),
			Entry(
				"all sub-assertions passed",
				NoneOf(pass, pass),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ none of (2 of the sub-assertions passed unexpectedly)`,
				`    ✓ <always pass>`,
				`    ✓ <always pass>`,
			),
			Entry(
				"some of the sub-assertions passed",
				NoneOf(pass, fail),
				false, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✗ none of (1 of the sub-assertions passed unexpectedly)`,
				`    ✓ <always pass>`,
				`    ✗ <always fail>`,
			),
			Entry(
				"none of the sub-assertions passed",
				NoneOf(fail, fail),
				true, // ok
				`--- ASSERTION REPORT ---`,
				``,
				`✓ none of`,
				`    ✗ <always fail>`,
				`    ✗ <always fail>`,
			),
		)
	})
})

type constAssertion bool

const (
	pass constAssertion = true
	fail constAssertion = false
)

func (a constAssertion) Begin(ExpectOptionSet) {}
func (a constAssertion) End()                  {}
func (a constAssertion) Ok() bool              { return bool(a) }
func (a constAssertion) Notify(fact.Fact)      {}
func (a constAssertion) BuildReport(ok bool, r render.Renderer) *Report {
	c := "<always fail>"
	if a {
		c = "<always pass>"
	}

	return &Report{
		TreeOk:   ok,
		Ok:       bool(a),
		Criteria: c,
	}
}
