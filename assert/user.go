package assert

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

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

	a.t.caller = callerName(0)
	a.assert(&a.t)
}

// Ok returns true if the assertion passed.
func (a *userAssertion) Ok() bool {
	return a.t.skipped || !a.t.failed
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

	if a.t.skipped {
		rep.Outcome = "the user-defined assertion was skipped"
	} else if a.t.failed {
		rep.Outcome = "the user-defined assertion failed"
	}

	if !verbose && (ok || a.Ok()) {
		return rep
	}

	rep.Explanation = a.t.explanation

	if len(a.t.messages) != 0 {
		s := rep.Section(logSection)

		for _, m := range a.t.messages {
			s.Content.WriteString(m)
			s.Content.WriteByte('\n')
		}
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

	name string

	m           sync.RWMutex
	skipped     bool
	failed      bool
	explanation string
	messages    []string
	cleanup     []func()
	caller      string
	helpers     map[string]struct{}
}

type abortUserAssertion struct{}

// Cleanup registers a function to be called when the test is complete. Cleanup
// functions will be called in last added, first called order.
func (t *T) Cleanup(fn func()) {
	t.cleanup = append(t.cleanup, fn)
}

// Error is equivalent to Log() followed by Fail().
func (t *T) Error(args ...interface{}) {
	t.Log(args...)
	t.fail("Error")
}

// Errorf is equivalent to Logf() followed by Fail().
func (t *T) Errorf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.fail("Errorf")
}

// Fail marks the function as having failed but continues execution.
func (t *T) Fail() {
	t.fail("Fail")
}

// FailNow marks the function as having failed and stops its execution.
func (t *T) FailNow() {
	t.fail("FailNow")
	panic(abortUserAssertion{})
}

// Failed reports whether the test has failed.
func (t *T) Failed() bool {
	return t.failed
}

// Fatal is equivalent to Log() followed by FailNow().
func (t *T) Fatal(args ...interface{}) {
	t.Log(args...)
	t.fail("Fatal")
	panic(abortUserAssertion{})
}

// Fatalf is equivalent to Logf() followed by FailNow().
func (t *T) Fatalf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.fail("Fatalf")
	panic(abortUserAssertion{})
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
	t.m.Lock()
	defer t.m.Unlock()

	if t.helpers == nil {
		t.helpers = map[string]struct{}{}
	}

	t.helpers[callerName(1)] = struct{}{}
}

// Log formats its arguments using default formatting, analogous to Println(),
// and records the text in the assertion report.
func (t *T) Log(args ...interface{}) {
	m := fmt.Sprint(args...)
	t.messages = append(t.messages, m)
}

// Logf formats its arguments according to the format, analogous to Printf(),
// and records the text in the assertion report.
func (t *T) Logf(format string, args ...interface{}) {
	m := fmt.Sprintf(format, args...)
	t.messages = append(t.messages, m)
}

// Name returns the name of the running test.
func (t *T) Name() string {
	// TODO: https://github.com/onsi/ginkgo/issues/582
	//
	// It would be good if we could get some more context here, but for the time
	// being we are keeping the testkit.T interface compatible with Ginkgo's
	// GinkgoTInterface, which does not have a Name() method.
	return t.name
}

// Skip is equivalent to Log() followed by SkipNow().
func (t *T) Skip(args ...interface{}) {
	t.Log(args...)
	t.skip("Skip")
}

// SkipNow marks the test as having been skipped and stops its execution.
func (t *T) SkipNow() {
	t.skip("SkipNow")
}

// Skipf is equivalent to Logf() followed by SkipNow().
func (t *T) Skipf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.skip("Skipf")
}

// Skipped reports whether the test was skipped.
func (t *T) Skipped() bool {
	return t.skipped
}

// skip marks the test as failed and sets the explanation message to m.
func (t *T) skip(fn string) {
	t.skipped = true
	t.explain(fn)
	panic(abortUserAssertion{})
}

// fail marks the test as failed and sets the explanation message to m.
func (t *T) fail(fn string) {
	t.failed = true
	t.explain(fn)
}

// explain populates t.explanation, including file/line information.
func (t *T) explain(fn string) {
	if t.explanation != "" {
		return
	}

	frame, direct := t.findFrame(3) // skip explain(), fail() / skip(), and their caller.

	file := "???"
	if frame.File != "" {
		file = filepath.Base(frame.File)
	}

	line := frame.Line
	if line == 0 {
		line = 1
	}

	if direct {
		t.explanation = fmt.Sprintf("%s() called at %s:%d", fn, file, line)
	} else {
		t.explanation = fmt.Sprintf("%s() called indirectly by call at %s:%d", fn, file, line)
	}
}

// findFrame searches, starting after skip frames, for the first caller frame
// in a function not marked as a helper and returns that frame.
//
// The search stops if it finds a method of userAssertion.
//
// It is assumed that t.m is already locked.
func (t *T) findFrame(skip int) (runtime.Frame, bool) {
	frames := stack(skip, 50)
	var first, prev runtime.Frame

	for {
		frame, more := frames.Next()
		if first.PC == 0 {
			first = frame
		}

		if frame.Function == t.caller {
			return prev, prev == first
		}

		_, isHelper := t.helpers[frame.Function]

		if !isHelper || !more {
			return frame, frame == first
		}

		prev = frame
	}
}

// callerName gives the function name (qualified with a package path)
// for the caller after skip frames (where 0 means the current function).
func callerName(skip int) string {
	frames := stack(skip, 1)
	frame, _ := frames.Next()
	return frame.Function
}

// stack returns a frames *above the caller* on the stack.
func stack(skip, max int) *runtime.Frames {
	var pc [50]uintptr

	// Add 3 extra frames to account for the caller, this function and
	// runtime.Callers() itself.
	n := runtime.Callers(skip+3, pc[:max])
	if n == 0 {
		panic("zero callers found")
	}

	return runtime.CallersFrames(pc[:n])
}
