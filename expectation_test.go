package testkit_test

import (
	"errors"

	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/fact"
)

var (
	// pass is an Expectation that is always passes.
	pass = staticExpectation{ok: true}

	// fail is an Expectation that always fails.
	fail = staticExpectation{ok: false}

	// failBeforeAction is an Expectation that fails before the Action occurs.
	failBeforeAction = staticExpectation{err: errors.New("<always fail before action>")}
)

// staticExpectation is is an Expectation that always produces the same result.
// It is intended to be used for testing the test system itself.
//
// It implements both the Expectation and Predicate interfaces.
type staticExpectation struct {
	ok  bool
	err error
}

func (e staticExpectation) Banner() string {
	if e.ok {
		return "TO [ALWAYS PASS]"
	}

	return "TO [ALWAYS FAIL]"
}

func (e staticExpectation) Predicate(PredicateOptions) (Predicate, error) {
	return e, e.err
}

func (e staticExpectation) Notify(fact.Fact) {}
func (e staticExpectation) Ok() bool         { return e.ok }
func (e staticExpectation) Done()            {}
func (e staticExpectation) Report(treeOk bool) *Report {
	c := "<always fail>"
	if e.ok {
		c = "<always pass>"
	}

	return &Report{
		TreeOk:   treeOk,
		Ok:       e.ok,
		Criteria: c,
	}
}

const (
	expectPass = true
	expectFail = false
)
