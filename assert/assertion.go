package assert

import (
	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// Operation is an enumeration of the test operations that can make assertions.
type Operation int

const (
	// ExecuteCommandOperation is an operation that makes assertions about what
	// happens when a command is executed.
	ExecuteCommandOperation Operation = iota

	// RecordEventOperation is an operation that makes assertions about what
	// happens when an event is recorded.
	RecordEventOperation

	// AdvanceTimeOperation is an operation that makes assertions about what
	// happens when the engine time advances.
	AdvanceTimeOperation

	// CallOperation is an operation that makes assertions about what happens
	// when a user-defined function is invoked.
	CallOperation
)

// An Assertion is a predicate for determining whether some specific criteria
// was met during a test.
type Assertion interface {
	fact.Observer

	// Begin is called to prepare the assertion for a new test.
	//
	// op is the operation that is making the assertion. c is the comparator
	// used to compare messages and other entities.
	Begin(op Operation, c compare.Comparator)

	// End is called once the test is complete.
	End()

	// Ok returns true if the assertion passed.
	Ok() bool

	// BuildReport generates a report about the assertion.
	//
	// ok is true if the assertion is considered to have passed. This may not be
	// the same value as returned from Ok() when this assertion is used as a
	// sub-assertion inside a composite.
	BuildReport(ok bool, r render.Renderer) *Report
}

// Nothing is an assertion that has no requirements.
var Nothing Assertion = nothingAssertion{}

type nothingAssertion struct{}

func (nothingAssertion) Notify(fact.Fact)                    {}
func (nothingAssertion) Begin(Operation, compare.Comparator) {}
func (nothingAssertion) End()                                {}
func (nothingAssertion) Ok() bool                            { return true }
func (nothingAssertion) BuildReport(ok bool, _ render.Renderer) *Report {
	return &Report{
		TreeOk:   ok,
		Ok:       true,
		Criteria: "no requirement",
	}
}
