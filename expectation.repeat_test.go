package testkit_test

import (
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/internal/x/xtesting"
)

func TestToRepeatedly(t *testing.T) {
	newFixture := func() (*testingmock.T, *Test) {
		mt := &testingmock.T{FailSilently: true}
		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "259ae495-fcef-43e2-986a-ea6b82f65fcd")
			},
		}
		return mt, Begin(mt, app)
	}

	cases := []struct {
		Name        string
		Expectation Expectation
		Passes      bool
		Report      reportMatcher
	}{
		{
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
		},
		{
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
		},
		{
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
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mt, tc := newFixture()
			tc.Expect(noop, c.Expectation)
			c.Report(mt)
			if mt.Failed() != !c.Passes {
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}
		})
	}

	t.Run("it produces the expected caption", func(t *testing.T) {
		mt, tc := newFixture()
		tc.Expect(
			noop,
			ToRepeatedly(
				"<description>",
				1,
				func(i int) Expectation {
					return pass
				},
			),
		)
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"--- expect [no-op] to <description> ---",
		)
	})

	t.Run("it panics if the description is empty", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			`ToRepeatedly("", 1, <func>): description must not be empty`,
			func() {
				ToRepeatedly("", 1, func(i int) Expectation { return nil })
			},
		)
	})

	t.Run("it panics if the count is zero", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			`ToRepeatedly("<description>", 0, <func>): n must be 1 or greater`,
			func() {
				ToRepeatedly("<description>", 0, func(i int) Expectation { return nil })
			},
		)
	})

	t.Run("it panics if the count is negative", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			`ToRepeatedly("<description>", -1, <func>): n must be 1 or greater`,
			func() {
				ToRepeatedly("<description>", -1, func(i int) Expectation { return nil })
			},
		)
	})

	t.Run("it panics if the function is nil", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			`ToRepeatedly("<description>", 1, <nil>): function must not be nil`,
			func() {
				ToRepeatedly("<description>", 1, nil)
			},
		)
	})
}
