package assert

import (
	"fmt"

	"github.com/dogmatiq/testkit/compare"
	"github.com/dogmatiq/testkit/engine/fact"
	"github.com/dogmatiq/testkit/render"
)

// Should returns an assertion that uses a user-defined function to check for
// specific criteria.
//
// cr is a human-readable description of the "criteria" or the "expectation"
// that results in a passing assertion. It should be phrased as an imperative
// statement, such as "insert a customer".
//
// fn is the function that performs the assertion logic. It is passed a *T
// value, which is a superset of the *testing.T struct that is passed
func Should(
	cr string,
	fn func(t *T),
) Assertion {
	return &userAssertion{
		t:      T{name: cr},
		assert: fn,
	}
}

// userAssertion is a user-defined assertion.
type userAssertion struct {
	t      T
	assert func(*T)
}

// Notify the observer of a fact.
func (a *userAssertion) Notify(f fact.Fact) {
	a.t.Facts = append(a.t.Facts, f)
}

// Begin is called to prepare the assertion for a new test.
//
// c is the comparator used to compare messages and other entities.
func (a *userAssertion) Begin(c compare.Comparator) {
	a.t.Comparator = c
}

// End is called once the test is complete.
func (a *userAssertion) End() {
	defer func() {
		for i := len(a.t.cleanup) - 1; i >= 0; i-- {
			a.t.cleanup[i]()
		}

		switch r := recover().(type) {
		case abortUserAssertion:
			return // keep to see coverage
		case nil:
			return // keep to see coverage
		default:
			panic(r)
		}
	}()

	a.assert(&a.t)
}

// Ok returns true if the assertion passed.
func (a *userAssertion) Ok() bool {
	return !a.t.failed
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be
// the same value as returned from Ok() when this assertion is used as a
// sub-assertion inside a composite.
func (a *userAssertion) BuildReport(ok, verbose bool, r render.Renderer) *Report {
	rep := &Report{
		TreeOk:   ok,
		Ok:       a.Ok(),
		Criteria: a.t.name,
	}

	if ok || a.Ok() {
		return rep
	}

	if a.t.failed {
		rep.Outcome = "the user-defined assertion failed"
	} else if a.t.skipped {
		rep.Outcome = "the user-defined assertion was skipped"
	}

	if len(a.t.messages) == 1 && a.t.explanation != "" {
		// If there is only one log message, and it was supplied by Error(),
		// Errorf(), Fatal(), or Fatalf(), show it as the explanation, rather
		// than in a logging section.
		rep.Explanation = a.t.explanation
		return rep
	}

	return rep
}

// T is a superset of *testing.T that is passed to user-defined assertion
// functions.
type T struct {
	// Comparator provides logic for comparing messages and application state.
	compare.Comparator

	// Facts is an ordered slice of the facts that occurred.
	Facts []fact.Fact

	name        string
	skipped     bool
	failed      bool
	explanation string
	messages    []string
	cleanup     []func()
}

type abortUserAssertion struct{}

// Cleanup registers a function to be called when the test is complete. Cleanup
// functions will be called in last added, first called order.
func (t *T) Cleanup(fn func()) {
	t.cleanup = append(t.cleanup, fn)
}

// Error is equivalent to Log() followed by Fail().
func (t *T) Error(args ...interface{}) {
	t.log(true, fmt.Sprint(args...))
	t.Fail()
}

// Errorf is equivalent to Logf() followed by Fail().
func (t *T) Errorf(format string, args ...interface{}) {
	t.log(true, fmt.Sprintf(format, args...))
	t.Fail()
}

// Fail marks the function as having failed but continues execution.
func (t *T) Fail() {
	t.failed = true
}

// FailNow marks the function as having failed and stops its execution.
func (t *T) FailNow() {
	t.failed = true
	panic(abortUserAssertion{})
}

// Failed reports whether the test has failed.
func (t *T) Failed() bool {
	return t.failed
}

// Fatal is equivalent to Log() followed by FailNow().
func (t *T) Fatal(args ...interface{}) {
	t.log(true, fmt.Sprint(args...))
	t.FailNow()
}

// Fatalf is equivalent to Logf() followed by FailNow().
func (t *T) Fatalf(format string, args ...interface{}) {
	t.log(true, fmt.Sprintf(format, args...))
	t.FailNow()
}

// Parallel signals that this test is to be run in parallel with (and only with)
// other parallel tests.
//
// It is a no-op in this implementation, but is included to increase
// compatibility with the *testing.T type.
func (t *T) Parallel() {
}

// Helper marks the calling function as a test helper function.
//
// It is a no-op in this implementation, but is included to increase
// compatibility with the *testing.T type.
func (t *T) Helper() {
}

// Log formats its arguments using default formatting, analogous to Println(),
// and records the text in the assertion report.
func (t *T) Log(args ...interface{}) {
	t.log(false, fmt.Sprint(args...))
}

// Logf formats its arguments according to the format, analogous to Printf(),
// and records the text in the assertion report.
func (t *T) Logf(format string, args ...interface{}) {
	t.log(false, fmt.Sprintf(format, args...))
}

// Name returns the name of the running test.
func (t *T) Name() string {
	return t.name
}

// Skip is equivalent to Log() followed by SkipNow().
func (t *T) Skip(args ...interface{}) {
	t.Log(args...)
	t.SkipNow()
}

// SkipNow marks the test as having been skipped and stops its execution.
func (t *T) SkipNow() {
	t.skipped = true
	panic(abortUserAssertion{})
}

// Skipf is equivalent to Logf() followed by SkipNow().
func (t *T) Skipf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.SkipNow()
}

// Skipped reports whether the test was skipped.
func (t *T) Skipped() bool {
	return t.skipped
}

func (t *T) log(fail bool, s string) {
	t.messages = append(t.messages, s)

	if fail && t.explanation == "" {
		t.explanation = s
	}
}
