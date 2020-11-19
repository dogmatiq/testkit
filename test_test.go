package testkit_test

import (
	"context"

	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/internal/testingmock"
	"github.com/dogmatiq/testkit/render"
	. "github.com/onsi/gomega"
)

var noop noopAction

// noopAction is an Action that does nothing. It is intended to be used for
// testing the test system itself.
type noopAction struct{}

func (noopAction) ExpectOptions() []ExpectOption                  { return nil }
func (noopAction) Apply(ctx context.Context, s ActionScope) error { return nil }

// staticExpectation is is an Expectation that always always produces the same
// result. It is intended to be used for testing the test system itself.
type staticExpectation bool

const (
	// pass is an Expectation that is always passes.
	pass staticExpectation = true

	// fail is an Expectation that always fails.
	fail staticExpectation = false
)

func (a staticExpectation) Begin(ExpectOptionSet) {}
func (a staticExpectation) End()                  {}
func (a staticExpectation) Ok() bool              { return bool(a) }
func (a staticExpectation) Notify(fact.Fact)      {}
func (a staticExpectation) BuildReport(ok bool, r render.Renderer) *assert.Report {
	c := "<always fail>"
	if a {
		c = "<always pass>"
	}

	return &assert.Report{
		TreeOk:   ok,
		Ok:       bool(a),
		Criteria: c,
	}
}

const (
	expectPass = true
	expectFail = false
)

// reportMatcher validates that some action produced a particular test report.
type reportMatcher func(*testingmock.T)

// expectReport is a helper function for testing that a testkit.Test produces the
// correct test report.
func expectReport(expected ...string) reportMatcher {
	// Always expect blank lines surrounding the report content.
	expected = append([]string{""}, expected...)
	expected = append(expected, "")

	return func(t *testingmock.T) {
		// Scan through the logs until we find of the start of the test report,
		// then assert that the remainder of the log content matches our
		// expectation.
		for i, l := range t.Logs {
			if l == "--- TEST REPORT ---" {
				Expect(t.Logs[i+1:]).To(Equal(expected))
				return
			}
		}

		// If we didn't find the test report at all just compare all of the logs
		// to the expectation so at least we know what *was* printed.
		Expect(t.Logs).To(Equal(expected))
	}
}
