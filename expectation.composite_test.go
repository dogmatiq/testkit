package testkit_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Context("composite expectations", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	g.BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "00df8612-2fd4-4ae3-9acf-afc2b4daf272")
			},
		}

		test = testkit.Begin(testingT, app)
	})

	testExpectationBehavior := func(
		e Expectation,
		ok bool,
		rm reportMatcher,
	) {
		test.Expect(noop, e)
		rm(testingT)
		gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
	}

	g.Describe("func AllOf()", func() {
		g.DescribeTable(
			"expectation behavior",
			testExpectationBehavior,
			g.Entry(
				"it flattens report output when there is a single child",
				AllOf(pass),
				expectPass,
				expectReport(
					`✓ <always pass>`,
				),
			),
			g.Entry(
				"it passes when all of the child expectations pass",
				AllOf(pass, pass),
				expectPass,
				expectReport(
					`✓ all of`,
					`    ✓ <always pass>`,
					`    ✓ <always pass>`,
				),
			),
			g.Entry(
				"it fails when some of the child expectations fail",
				AllOf(pass, fail),
				expectFail,
				expectReport(
					`✗ all of (1 of the expectations failed)`,
					`    ✓ <always pass>`,
					`    ✗ <always fail>`,
				),
			),
			g.Entry(
				"it fails when all of the child expectations fail",
				AllOf(fail, fail),
				expectFail,
				expectReport(
					`✗ all of (2 of the expectations failed)`,
					`    ✗ <always fail>`,
					`    ✗ <always fail>`,
				),
			),
		)

		g.It("produces the expected caption", func() {
			test.Expect(
				noop,
				AllOf(pass, fail),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement(
				"--- expect [no-op] to meet 2 expectations ---",
			))
		})

		g.It("fails the test if one of its children cannot construct a predicate", func() {
			test.Expect(
				noop,
				AllOf(pass, failBeforeAction),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement("<always fail before action>"))
			gm.Expect(testingT.Failed()).To(gm.BeTrue())
		})

		g.It("panics if no children are provided", func() {
			gm.Expect(func() {
				AllOf()
			}).To(gm.PanicWith("AllOf(): at least one child expectation must be provided"))
		})
	})

	g.Describe("func AnyOf()", func() {
		g.DescribeTable(
			"expectation behavior",
			testExpectationBehavior,
			g.Entry(
				"it flattens report output when there is a single child",
				AnyOf(pass),
				expectPass,
				expectReport(
					`✓ <always pass>`,
				),
			),
			g.Entry(
				"it passes when all of the child expectations pass",
				AnyOf(pass, pass),
				expectPass,
				expectReport(
					`✓ any of`,
					`    ✓ <always pass>`,
					`    ✓ <always pass>`,
				),
			),
			g.Entry(
				"it passes when some of the child expectations fail",
				AnyOf(pass, fail),
				expectPass,
				expectReport(
					`✓ any of`,
					`    ✓ <always pass>`,
					`    ✗ <always fail>`,
				),
			),
			g.Entry(
				"it fails when all of the child expectations fail",
				AnyOf(fail, fail),
				expectFail,
				expectReport(
					`✗ any of (all 2 of the expectations failed)`,
					`    ✗ <always fail>`,
					`    ✗ <always fail>`,
				),
			),
		)

		g.It("produces the expected caption", func() {
			test.Expect(
				noop,
				AnyOf(pass, fail),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement(
				"--- expect [no-op] to meet at least one of 2 expectations ---",
			))
		})

		g.It("fails the test if one of its children cannot construct a predicate", func() {
			test.Expect(
				noop,
				AnyOf(pass, failBeforeAction),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement("<always fail before action>"))
			gm.Expect(testingT.Failed()).To(gm.BeTrue())
		})

		g.It("panics if no children are provided", func() {
			gm.Expect(func() {
				AnyOf()
			}).To(gm.PanicWith("AnyOf(): at least one child expectation must be provided"))
		})
	})

	g.Describe("func NoneOf()", func() {
		g.DescribeTable(
			"expectation behavior",
			testExpectationBehavior,
			g.Entry(
				"it does not flatten report output when there is a single child",
				NoneOf(pass),
				expectFail,
				expectReport(
					`✗ none of (the expectation passed unexpectedly)`,
					`    ✓ <always pass>`,
				),
			),
			g.Entry(
				"it fails when all of the child expecations pass",
				NoneOf(pass, pass),
				expectFail,
				expectReport(
					`✗ none of (2 of the expectations passed unexpectedly)`,
					`    ✓ <always pass>`,
					`    ✓ <always pass>`,
				),
			),
			g.Entry(
				"it fails when some of the child expectations pass",
				NoneOf(pass, fail),
				expectFail,
				expectReport(
					`✗ none of (1 of the expectations passed unexpectedly)`,
					`    ✓ <always pass>`,
					`    ✗ <always fail>`,
				),
			),
			g.Entry(
				"passes when all of the child expectations fail",
				NoneOf(fail, fail),
				expectPass,
				expectReport(
					`✓ none of`,
					`    ✗ <always fail>`,
					`    ✗ <always fail>`,
				),
			),
		)

		g.It("produces the expected caption", func() {
			test.Expect(
				noop,
				NoneOf(pass, fail),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement(
				"--- expect [no-op] not to meet any of 2 expectations ---",
			))
		})

		g.It("produces the expected caption when there is only one child", func() {
			test.Expect(
				noop,
				NoneOf(pass),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement(
				"--- expect [no-op] not to [always pass] ---",
			))
		})

		g.It("fails the test if one of its children cannot construct a predicate", func() {
			test.Expect(
				noop,
				NoneOf(pass, failBeforeAction),
			)

			gm.Expect(testingT.Logs).To(gm.ContainElement("<always fail before action>"))
			gm.Expect(testingT.Failed()).To(gm.BeTrue())
		})

		g.It("panics if no children are provided", func() {
			gm.Expect(func() {
				NoneOf()
			}).To(gm.PanicWith("NoneOf(): at least one child expectation must be provided"))
		})
	})
})
