package testkit_test

import (
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/assert"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

const (
	// pass is an Expectation that is always passes.
	pass staticExpectation = true

	// fail is an Expectation that always fails.
	fail staticExpectation = false
)

// staticExpectation is is an Expectation that always produces the same result.
// It is intended to be used for testing the test system itself.
type staticExpectation bool

func (a staticExpectation) Banner() string {
	if a {
		return "TO [ALWAYS PASS]"
	}

	return "TO [ALWAYS FAIL]"
}

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
