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

var _ = g.Describe("func ToRepeatedly()", func() {
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
				c.Identity("<app>", "259ae495-fcef-43e2-986a-ea6b82f65fcd")
			},
		}

		test = testkit.Begin(testingT, app)
	})

	g.DescribeTable(
		"expectation behavior",
		func(
			e Expectation,
			ok bool,
			rm reportMatcher,
		) {
			test.Expect(noop, e)
			rm(testingT)
			gm.Expect(testingT.Failed()).To(gm.Equal(!ok))
		},
		g.Entry(
			"it passes when all of the repeated expectations pass",
			ToRepeatedly(
				"<description>",
				2,
				func(i int) Expectation {
					switch i {
					case 0:
						return pass
					case 1:
						return pass
					default:
						panic("unexpected index")
					}
				},
			),
			expectPass,
			expectReport(
				`✓ <description>`,
			),
		),
		g.Entry(
			"it fails when any of the repeated expectations fail",
			ToRepeatedly(
				"<description>",
				2,
				func(i int) Expectation {
					switch i {
					case 0:
						return pass
					case 1:
						return fail
					default:
						panic("unexpected index")
					}
				},
			),
			expectFail,
			expectReport(
				`✗ <description> (1 of 2 iteration(s) failed, iteration #1 shown)`,
				`    ✗ <always fail>`,
			),
		),
		g.Entry(
			"it fails when all of the repeated expectations fail",
			ToRepeatedly(
				"<description>",
				2,
				func(i int) Expectation {
					switch i {
					case 0:
						return fail
					case 1:
						return fail
					default:
						panic("unexpected index")
					}
				},
			),
			expectFail,
			expectReport(
				`✗ <description> (2 of 2 iteration(s) failed, iteration #0 shown)`,
				`    ✗ <always fail>`,
			),
		),
	)

	g.It("produces the expected caption", func() {
		test.Expect(
			noop,
			ToRepeatedly(
				"<description>",
				1,
				func(i int) Expectation {
					return pass
				},
			),
		)

		gm.Expect(testingT.Logs).To(gm.ContainElement(
			"--- expect [no-op] to <description> ---",
		))
	})

	g.It("panics if the description is empty", func() {
		gm.Expect(func() {
			ToRepeatedly("", 1, func(i int) Expectation { return nil })
		}).To(gm.PanicWith(`ToRepeatedly("", 1, <func>): description must not be empty`))
	})

	g.It("panics if the count is zero", func() {
		gm.Expect(func() {
			ToRepeatedly("<description>", 0, func(i int) Expectation { return nil })
		}).To(gm.PanicWith(`ToRepeatedly("<description>", 0, <func>): n must be 1 or greater`))
	})

	g.It("panics if the count is negative", func() {
		gm.Expect(func() {
			ToRepeatedly("<description>", -1, func(i int) Expectation { return nil })
		}).To(gm.PanicWith(`ToRepeatedly("<description>", -1, <func>): n must be 1 or greater`))
	})

	g.It("panics if the function is nil", func() {
		gm.Expect(func() {
			ToRepeatedly("<description>", 1, nil)
		}).To(gm.PanicWith(`ToRepeatedly("<description>", 1, <nil>): function must not be nil`))
	})
})
