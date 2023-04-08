package testkit_test

import (
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	"github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Context("composite expectations", func() {
	var (
		testingT *testingmock.T
		app      dogma.Application
		test     *Test
	)

	BeforeEach(func() {
		testingT = &testingmock.T{
			FailSilently: true,
		}

		app = &Application{
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
		Expect(testingT.Failed()).To(Equal(!ok))
	}

	Describe("func AllOf()", func() {
		DescribeTable(
			"expectation behavior",
			testExpectationBehavior,
			Entry(
				"it flattens report output when there is a single child",
				AllOf(pass),
				expectPass,
				expectReport(
					`✓ <always pass>`,
				),
			),
			Entry(
				"it passes when all of the child expectations pass",
				AllOf(pass, pass),
				expectPass,
				expectReport(
					`✓ all of`,
					`    ✓ <always pass>`,
					`    ✓ <always pass>`,
				),
			),
			Entry(
				"it fails when some of the child expectations fail",
				AllOf(pass, fail),
				expectFail,
				expectReport(
					`✗ all of (1 of the expectations failed)`,
					`    ✓ <always pass>`,
					`    ✗ <always fail>`,
				),
			),
			Entry(
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

		It("produces the expected caption", func() {
			test.Expect(
				noop,
				AllOf(pass, fail),
			)

			Expect(testingT.Logs).To(ContainElement(
				"--- expect [no-op] to meet 2 expectations ---",
			))
		})

		It("fails the test if one of its children can not construct a predicate", func() {
			test.Expect(
				noop,
				AllOf(pass, failBeforeAction),
			)

			Expect(testingT.Logs).To(ContainElement("<always fail before action>"))
			Expect(testingT.Failed()).To(BeTrue())
		})

		It("panics if no children are provided", func() {
			Expect(func() {
				AllOf()
			}).To(PanicWith("AllOf(): at least one child expectation must be provided"))
		})
	})

	Describe("func AnyOf()", func() {
		DescribeTable(
			"expectation behavior",
			testExpectationBehavior,
			Entry(
				"it flattens report output when there is a single child",
				AnyOf(pass),
				expectPass,
				expectReport(
					`✓ <always pass>`,
				),
			),
			Entry(
				"it passes when all of the child expectations pass",
				AnyOf(pass, pass),
				expectPass,
				expectReport(
					`✓ any of`,
					`    ✓ <always pass>`,
					`    ✓ <always pass>`,
				),
			),
			Entry(
				"it passes when some of the child expectations fail",
				AnyOf(pass, fail),
				expectPass,
				expectReport(
					`✓ any of`,
					`    ✓ <always pass>`,
					`    ✗ <always fail>`,
				),
			),
			Entry(
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

		It("produces the expected caption", func() {
			test.Expect(
				noop,
				AnyOf(pass, fail),
			)

			Expect(testingT.Logs).To(ContainElement(
				"--- expect [no-op] to meet at least one of 2 expectations ---",
			))
		})

		It("fails the test if one of its children can not construct a predicate", func() {
			test.Expect(
				noop,
				AnyOf(pass, failBeforeAction),
			)

			Expect(testingT.Logs).To(ContainElement("<always fail before action>"))
			Expect(testingT.Failed()).To(BeTrue())
		})

		It("panics if no children are provided", func() {
			Expect(func() {
				AnyOf()
			}).To(PanicWith("AnyOf(): at least one child expectation must be provided"))
		})
	})

	Describe("func NoneOf()", func() {
		DescribeTable(
			"expectation behavior",
			testExpectationBehavior,
			Entry(
				"it does not flatten report output when there is a single child",
				NoneOf(pass),
				expectFail,
				expectReport(
					`✗ none of (the expectation passed unexpectedly)`,
					`    ✓ <always pass>`,
				),
			),
			Entry(
				"it fails when all of the child expecations pass",
				NoneOf(pass, pass),
				expectFail,
				expectReport(
					`✗ none of (2 of the expectations passed unexpectedly)`,
					`    ✓ <always pass>`,
					`    ✓ <always pass>`,
				),
			),
			Entry(
				"it fails when some of the child expectations pass",
				NoneOf(pass, fail),
				expectFail,
				expectReport(
					`✗ none of (1 of the expectations passed unexpectedly)`,
					`    ✓ <always pass>`,
					`    ✗ <always fail>`,
				),
			),
			Entry(
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

		It("produces the expected caption", func() {
			test.Expect(
				noop,
				NoneOf(pass, fail),
			)

			Expect(testingT.Logs).To(ContainElement(
				"--- expect [no-op] not to meet any of 2 expectations ---",
			))
		})

		It("produces the expected caption when there is only one child", func() {
			test.Expect(
				noop,
				NoneOf(pass),
			)

			Expect(testingT.Logs).To(ContainElement(
				"--- expect [no-op] not to [always pass] ---",
			))
		})

		It("fails the test if one of its children can not construct a predicate", func() {
			test.Expect(
				noop,
				NoneOf(pass, failBeforeAction),
			)

			Expect(testingT.Logs).To(ContainElement("<always fail before action>"))
			Expect(testingT.Failed()).To(BeTrue())
		})

		It("panics if no children are provided", func() {
			Expect(func() {
				NoneOf()
			}).To(PanicWith("NoneOf(): at least one child expectation must be provided"))
		})
	})
})
