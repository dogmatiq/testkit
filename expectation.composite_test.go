package testkit_test

import (
	"testing"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/x/xtesting"
)

func newCompositeFixture() (*testingmock.T, *Test) {
	mt := &testingmock.T{FailSilently: true}
	app := &ApplicationStub{
		ConfigureFunc: func(c dogma.ApplicationConfigurer) {
			c.Identity("<app>", "00df8612-2fd4-4ae3-9acf-afc2b4daf272")
		},
	}
	return mt, Begin(mt, app)
}

func TestAllOf(t *testing.T) {
	cases := []struct {
		Name        string
		Expectation Expectation
		Passes      bool
		Report      reportMatcher
	}{
		{
			"it flattens report output when there is a single child",
			AllOf(pass),
			expectPass,
			expectReport(
				`✓ <always pass>`,
			),
		},
		{
			"it passes when all of the child expectations pass",
			AllOf(pass, pass),
			expectPass,
			expectReport(
				`✓ all of`,
				`    ✓ <always pass>`,
				`    ✓ <always pass>`,
			),
		},
		{
			"it fails when some of the child expectations fail",
			AllOf(pass, fail),
			expectFail,
			expectReport(
				`✗ all of (1 of the expectations failed)`,
				`    ✓ <always pass>`,
				`    ✗ <always fail>`,
			),
		},
		{
			"it fails when all of the child expectations fail",
			AllOf(fail, fail),
			expectFail,
			expectReport(
				`✗ all of (2 of the expectations failed)`,
				`    ✗ <always fail>`,
				`    ✗ <always fail>`,
			),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			mt, tc := newCompositeFixture()
			tc.Expect(noop, c.Expectation)
			c.Report(mt)
			if mt.Failed() != !c.Passes {
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}
		})
	}

	t.Run("it produces the expected caption", func(t *testing.T) {
		mt, tc := newCompositeFixture()
		tc.Expect(noop, AllOf(pass, fail))
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"--- expect [no-op] to meet 2 expectations ---",
		)
	})

	t.Run("it fails the test if one of its children cannot construct a predicate", func(t *testing.T) {
		mt, tc := newCompositeFixture()
		tc.Expect(noop, AllOf(pass, failBeforeAction))
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"<always fail before action>",
		)
		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
	})

	t.Run("it panics if no children are provided", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"AllOf(): at least one child expectation must be provided",
			func() { AllOf() },
		)
	})
}

func TestAnyOf(t *testing.T) {
	cases := []struct {
		Name        string
		Expectation Expectation
		Passes      bool
		Report      reportMatcher
	}{
		{
			"it flattens report output when there is a single child",
			AnyOf(pass),
			expectPass,
			expectReport(
				`✓ <always pass>`,
			),
		},
		{
			"it passes when all of the child expectations pass",
			AnyOf(pass, pass),
			expectPass,
			expectReport(
				`✓ any of`,
				`    ✓ <always pass>`,
				`    ✓ <always pass>`,
			),
		},
		{
			"it passes when some of the child expectations fail",
			AnyOf(pass, fail),
			expectPass,
			expectReport(
				`✓ any of`,
				`    ✓ <always pass>`,
				`    ✗ <always fail>`,
			),
		},
		{
			"it fails when all of the child expectations fail",
			AnyOf(fail, fail),
			expectFail,
			expectReport(
				`✗ any of (all 2 of the expectations failed)`,
				`    ✗ <always fail>`,
				`    ✗ <always fail>`,
			),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			mt, tc := newCompositeFixture()
			tc.Expect(noop, c.Expectation)
			c.Report(mt)
			if mt.Failed() != !c.Passes {
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}
		})
	}

	t.Run("it produces the expected caption", func(t *testing.T) {
		mt, tc := newCompositeFixture()
		tc.Expect(noop, AnyOf(pass, fail))
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"--- expect [no-op] to meet at least one of 2 expectations ---",
		)
	})

	t.Run("it fails the test if one of its children cannot construct a predicate", func(t *testing.T) {
		mt, tc := newCompositeFixture()
		tc.Expect(noop, AnyOf(pass, failBeforeAction))
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"<always fail before action>",
		)
		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
	})

	t.Run("it panics if no children are provided", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"AnyOf(): at least one child expectation must be provided",
			func() { AnyOf() },
		)
	})
}

func TestNoneOf(t *testing.T) {
	cases := []struct {
		Name        string
		Expectation Expectation
		Passes      bool
		Report      reportMatcher
	}{
		{
			"it does not flatten report output when there is a single child",
			NoneOf(pass),
			expectFail,
			expectReport(
				`✗ none of (the expectation passed unexpectedly)`,
				`    ✓ <always pass>`,
			),
		},
		{
			"it fails when all of the child expecations pass",
			NoneOf(pass, pass),
			expectFail,
			expectReport(
				`✗ none of (2 of the expectations passed unexpectedly)`,
				`    ✓ <always pass>`,
				`    ✓ <always pass>`,
			),
		},
		{
			"it fails when some of the child expectations pass",
			NoneOf(pass, fail),
			expectFail,
			expectReport(
				`✗ none of (1 of the expectations passed unexpectedly)`,
				`    ✓ <always pass>`,
				`    ✗ <always fail>`,
			),
		},
		{
			"passes when all of the child expectations fail",
			NoneOf(fail, fail),
			expectPass,
			expectReport(
				`✓ none of`,
				`    ✗ <always fail>`,
				`    ✗ <always fail>`,
			),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			mt, tc := newCompositeFixture()
			tc.Expect(noop, c.Expectation)
			c.Report(mt)
			if mt.Failed() != !c.Passes {
				t.Fatalf("testingT.Failed() = %v, want %v", mt.Failed(), !c.Passes)
			}
		})
	}

	t.Run("it produces the expected caption", func(t *testing.T) {
		mt, tc := newCompositeFixture()
		tc.Expect(noop, NoneOf(pass, fail))
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"--- expect [no-op] not to meet any of 2 expectations ---",
		)
	})

	t.Run("it produces the expected caption when there is only one child", func(t *testing.T) {
		mt, tc := newCompositeFixture()
		tc.Expect(noop, NoneOf(pass))
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"--- expect [no-op] not to [always pass] ---",
		)
	})

	t.Run("it fails the test if one of its children cannot construct a predicate", func(t *testing.T) {
		mt, tc := newCompositeFixture()
		tc.Expect(noop, NoneOf(pass, failBeforeAction))
		xtesting.ExpectContains[string](
			t,
			"expected log message not found",
			mt.Logs,
			"<always fail before action>",
		)
		if !mt.Failed() {
			t.Fatal("expected test to fail")
		}
	})

	t.Run("it panics if no children are provided", func(t *testing.T) {
		xtesting.ExpectPanic(
			t,
			"NoneOf(): at least one child expectation must be provided",
			func() { NoneOf() },
		)
	})
}
