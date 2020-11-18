package assert

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

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
// fn is the function that performs the assertion logic. It is passed a *S
// value, which is analogous to Go's *testing.T, and provides an almost
// identical interface.
func Should(
	cr string,
	fn func(s *S),
) Assertion {
	return &userAssertion{
		s:      S{name: cr},
		assert: fn,
	}
}

// userAssertion is a user-defined assertion.
type userAssertion struct {
	s      S
	assert func(*S)
}

// Notify the observer of a fact.
func (a *userAssertion) Notify(f fact.Fact) {
	a.s.Facts = append(a.s.Facts, f)
}

// Begin is called to prepare the assertion for a new test.
func (a *userAssertion) Begin(o ExpectOptionSet) {
	a.s.Options = o
}

// End is called once the test is complete.
func (a *userAssertion) End() {
	defer func() {
		for i := len(a.s.cleanup) - 1; i >= 0; i-- {
			a.s.cleanup[i]()
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

	a.s.caller = callerName(0)
	a.assert(&a.s)
}

// Ok returns true if the assertion passed.
func (a *userAssertion) Ok() bool {
	return a.s.skipped || !a.s.failed
}

// BuildReport generates a report about the assertion.
//
// ok is true if the assertion is considered to have passed. This may not be
// the same value as returned from Ok() when this assertion is used as a
// sub-assertion inside a composite.
func (a *userAssertion) BuildReport(ok bool, r render.Renderer) *Report {
	rep := &Report{
		TreeOk:   ok,
		Ok:       a.Ok(),
		Criteria: a.s.name,
	}

	if a.s.skipped {
		rep.Outcome = "the user-defined assertion was skipped"
	} else if a.s.failed {
		rep.Outcome = "the user-defined assertion failed"
	}

	rep.Explanation = a.s.explanation

	if len(a.s.messages) != 0 {
		s := rep.Section(logSection)

		for _, m := range a.s.messages {
			s.Content.WriteString(m)
			s.Content.WriteByte('\n')
		}
	}

	return rep
}

// S is passed to assertions made via Should() to allow the user-defined
// criteria to be enforced. It is analogous the *testing.T type that is passed
// to tests in the native Go test runner.
type S struct {
	// Options contains the set of options that dictate the behavior of the
	// expectation.
	Options ExpectOptionSet

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
func (s *S) Cleanup(fn func()) {
	s.m.Lock()
	defer s.m.Unlock()

	s.cleanup = append(s.cleanup, fn)
}

// Error is equivalent to Log() followed by Fail().
func (s *S) Error(args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.log(args)
	s.fail("Error")
}

// Errorf is equivalent to Logf() followed by Fail().
func (s *S) Errorf(format string, args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.logf(format, args)
	s.fail("Errorf")
}

// Fail marks the function as having failed but continues execution.
func (s *S) Fail() {
	s.m.Lock()
	defer s.m.Unlock()

	s.fail("Fail")
}

// FailNow marks the function as having failed and stops its execution.
func (s *S) FailNow() {
	s.m.Lock()
	defer s.m.Unlock()

	s.fail("FailNow")
	panic(abortUserAssertion{})
}

// Failed reports whether the test has failed.
func (s *S) Failed() bool {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.failed
}

// Fatal is equivalent to Log() followed by FailNow().
func (s *S) Fatal(args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.log(args)
	s.fail("Fatal")
	panic(abortUserAssertion{})
}

// Fatalf is equivalent to Logf() followed by FailNow().
func (s *S) Fatalf(format string, args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.logf(format, args)
	s.fail("Fatalf")
	panic(abortUserAssertion{})
}

// Parallel signals that this test is to be run in parallel with (and only with)
// other parallel tests.
//
// It is a no-op in this implementation, but is included to increase
// compatibility with the *testing.T type.
func (s *S) Parallel() {
}

// Helper marks the calling function as a test helper function.
func (s *S) Helper() {
	s.m.Lock()
	defer s.m.Unlock()

	if s.helpers == nil {
		s.helpers = map[string]struct{}{}
	}

	s.helpers[callerName(1)] = struct{}{}
}

// Log formats its arguments using default formatting, analogous to Println(),
// and records the text in the assertion report.
func (s *S) Log(args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.log(args)
}

// Logf formats its arguments according to the format, analogous to Printf(),
// and records the text in the assertion report.
func (s *S) Logf(format string, args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.logf(format, args)
}

// Name returns the name of the running test.
func (s *S) Name() string {
	// TODO: https://github.com/onsi/ginkgo/issues/582
	//
	// It would be good if we could get some more context here, but for the time
	// being we are keeping the testkit.T interface compatible with Ginkgo's
	// GinkgoTInterface, which does not have a Name() method.
	return s.name
}

// Skip is equivalent to Log() followed by SkipNow().
func (s *S) Skip(args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.log(args)
	s.skip("Skip")
}

// SkipNow marks the test as having been skipped and stops its execution.
func (s *S) SkipNow() {
	s.m.Lock()
	defer s.m.Unlock()

	s.skip("SkipNow")
}

// Skipf is equivalent to Logf() followed by SkipNow().
func (s *S) Skipf(format string, args ...interface{}) {
	s.m.Lock()
	defer s.m.Unlock()

	s.logf(format, args)
	s.skip("Skipf")
}

// Skipped reports whether the test was skipped.
func (s *S) Skipped() bool {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.skipped
}

// log adds a log message.
func (s *S) log(args []interface{}) {
	m := fmt.Sprint(args...)
	s.messages = append(s.messages, m)
}

// logf formats and adds a log message.
func (s *S) logf(format string, args []interface{}) {
	m := fmt.Sprintf(format, args...)
	s.messages = append(s.messages, m)
}

// skip marks the test as failed and sets the explanation message to m.
func (s *S) skip(fn string) {
	s.skipped = true
	s.explain(fn)
	panic(abortUserAssertion{})
}

// fail marks the test as failed and sets the explanation message to m.
func (s *S) fail(fn string) {
	s.failed = true
	s.explain(fn)
}

// explain populates t.explanation, including file/line information.
func (s *S) explain(fn string) {
	if s.explanation != "" {
		return
	}

	frame, direct := s.findFrame(3) // skip explain(), fail() / skip(), and their caller.

	file := "???"
	if frame.File != "" {
		file = filepath.Base(frame.File)
	}

	line := frame.Line
	if line == 0 {
		line = 1
	}

	if direct {
		s.explanation = fmt.Sprintf("%s() called at %s:%d", fn, file, line)
	} else {
		s.explanation = fmt.Sprintf("%s() called indirectly by call at %s:%d", fn, file, line)
	}
}

// findFrame searches, starting after skip frames, for the first caller frame
// in a function not marked as a helper and returns that frame.
//
// The search stops if it finds a method of userAssertion.
//
// It is assumed that s.m is already locked.
func (s *S) findFrame(skip int) (runtime.Frame, bool) {
	frames := stack(skip, 50)
	var first, prev runtime.Frame

	for {
		frame, more := frames.Next()
		if first.PC == 0 {
			first = frame
		}

		if frame.Function == s.caller {
			return prev, prev == first
		}

		_, isHelper := s.helpers[frame.Function]

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
