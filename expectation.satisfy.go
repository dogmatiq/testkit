package testkit

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/dogmatiq/testkit/fact"
)

// ToSatisfy returns an expectation that calls a function to check for arbitrary
// criteria.
//
// desc is a human-readable description of the expectation. It should be phrased
// as an imperative statement, such as "debit the customer".
//
// pred is a function that performs the expectation logic. It is passed a
// *SatisfyT, which is analogous to Go's *testing.T, and provides an almost
// identical interface.
func ToSatisfy(
	desc string,
	pred func(*SatisfyT),
) Expectation {
	if pred == nil {
		panic(fmt.Sprintf("ToSatisfy(%#v, <nil>): function must not be nil", desc))
	}

	if desc == "" {
		panic(fmt.Sprintf("ToSatisfy(%#v, <func>): description must not be empty", desc))
	}

	return &satisfyExpectation{
		criteria: desc,
		pred:     pred,
	}
}

// satisfyExpectation is an Expectation that calls a user-supplied function to
// check for arbitrary criteria.
//
// It is the implementation used by ToSatisfy().
type satisfyExpectation struct {
	criteria string
	pred     func(*SatisfyT)
}

func (e *satisfyExpectation) Caption() string {
	return "to " + e.criteria
}

func (e satisfyExpectation) Predicate(s PredicateScope) (Predicate, error) {
	return &satisfyPredicate{
		criteria: e.criteria,
		pred:     e.pred,
		satisfyT: SatisfyT{
			Options: s.Options,
			name:    e.criteria,
		},
	}, nil
}

// compositePredicate is the Predicate implementation for satisfyExpectation.
type satisfyPredicate struct {
	criteria string
	pred     func(*SatisfyT)
	satisfyT SatisfyT
}

func (p *satisfyPredicate) Notify(f fact.Fact) {
	p.satisfyT.Facts = append(p.satisfyT.Facts, f)
}

func (p *satisfyPredicate) Ok() bool {
	return p.satisfyT.skipped || !p.satisfyT.failed
}

func (p *satisfyPredicate) Done() {
	defer func() {
		switch r := recover().(type) {
		case abortSentinel:
			return // keep to see coverage
		case nil:
			return // keep to see coverage
		default:
			panic(r)
		}
	}()

	defer p.satisfyT.close()

	p.satisfyT.caller = callerName(0)
	p.pred(&p.satisfyT)
}

func (p *satisfyPredicate) Report(treeOk bool) *Report {
	rep := &Report{
		TreeOk:   treeOk,
		Ok:       p.Ok(),
		Criteria: p.criteria,
	}

	if p.satisfyT.skipped {
		rep.Outcome = "the expectation was skipped"
	} else if p.satisfyT.failed {
		rep.Outcome = "the expectation failed"
	}

	rep.Explanation = p.satisfyT.explanation

	if len(p.satisfyT.messages) != 0 {
		s := rep.Section(logSection)

		for _, m := range p.satisfyT.messages {
			s.Content.WriteString(m)
			s.Content.WriteByte('\n')
		}
	}

	return rep
}

// SatisfyT is used within expectations made via ToSatisfy() to enforce the
// expectation.
//
// It is analogous to the *testing.T type that is passed to tests in the native
// Go test runner.
type SatisfyT struct {
	// Options contains the set of options that change the behavior of the
	// predicate function.
	//
	// They are not necessarily applicable to every predicate.
	Options PredicateOptions

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

// abortSentinel is a panic value used to detect when execution of an
// expectation function has been aborted early by SatisfyT.FailNow() and similar
// methods.
type abortSentinel struct{}

// Cleanup registers a function to be called when the test is complete. Cleanup
// functions will be called in last added, first called order.
func (t *SatisfyT) Cleanup(fn func()) {
	t.m.Lock()
	defer t.m.Unlock()

	t.cleanup = append(t.cleanup, fn)
}

// Error is equivalent to Log() followed by Fail().
func (t *SatisfyT) Error(args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.log(args)
	t.fail("Error")
}

// Errorf is equivalent to Logf() followed by Fail().
func (t *SatisfyT) Errorf(format string, args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.logf(format, args)
	t.fail("Errorf")
}

// Fail marks the function as having failed but continues execution.
func (t *SatisfyT) Fail() {
	t.m.Lock()
	defer t.m.Unlock()

	t.fail("Fail")
}

// FailNow marks the function as having failed and stops its execution.
func (t *SatisfyT) FailNow() {
	t.m.Lock()
	defer t.m.Unlock()

	t.fail("FailNow")
	panic(abortSentinel{})
}

// Failed reports whether the test has failed.
func (t *SatisfyT) Failed() bool {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.failed
}

// Fatal is equivalent to Log() followed by FailNow().
func (t *SatisfyT) Fatal(args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.log(args)
	t.fail("Fatal")
	panic(abortSentinel{})
}

// Fatalf is equivalent to Logf() followed by FailNow().
func (t *SatisfyT) Fatalf(format string, args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.logf(format, args)
	t.fail("Fatalf")
	panic(abortSentinel{})
}

// Parallel signals that this test is to be run in parallel with (and only with)
// other parallel tests.
//
// It is a no-op in this implementation, but is included to increase
// compatibility with the *testing.T type.
func (t *SatisfyT) Parallel() {
}

// Helper marks the calling function as a test helper function.
func (t *SatisfyT) Helper() {
	t.m.Lock()
	defer t.m.Unlock()

	if t.helpers == nil {
		t.helpers = map[string]struct{}{}
	}

	t.helpers[callerName(1)] = struct{}{}
}

// Log formats its arguments using default formatting, analogous to Println(),
// and records the text in the test report.
func (t *SatisfyT) Log(args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.log(args)
}

// Logf formats its arguments according to the format, analogous to Printf(),
// and records the text in the test report.
func (t *SatisfyT) Logf(format string, args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.logf(format, args)
}

// Name returns the name of the running test.
func (t *SatisfyT) Name() string {
	// TODO: https://github.com/onsi/ginkgo/issues/582
	//
	// It would be good if we could get some more context here, but for the time
	// being we are keeping the testkit.T interface compatible with Ginkgo's
	// GinkgoTInterface, which does not have a Name() method.
	return t.name
}

// Skip is equivalent to Log() followed by SkipNow().
func (t *SatisfyT) Skip(args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.log(args)
	t.skip("Skip")
}

// SkipNow marks the test as having been skipped and stops its execution.
func (t *SatisfyT) SkipNow() {
	t.m.Lock()
	defer t.m.Unlock()

	t.skip("SkipNow")
}

// Skipf is equivalent to Logf() followed by SkipNow().
func (t *SatisfyT) Skipf(format string, args ...interface{}) {
	t.m.Lock()
	defer t.m.Unlock()

	t.logf(format, args)
	t.skip("Skipf")
}

// Skipped reports whether the test was skipped.
func (t *SatisfyT) Skipped() bool {
	t.m.RLock()
	defer t.m.RUnlock()

	return t.skipped
}

// log adds a log message.
func (t *SatisfyT) log(args []interface{}) {
	m := fmt.Sprint(args...)
	t.messages = append(t.messages, m)
}

// logf formats and adds a log message.
func (t *SatisfyT) logf(format string, args []interface{}) {
	m := fmt.Sprintf(format, args...)
	t.messages = append(t.messages, m)
}

// skip marks the test as skipped.
// fn is the name of the function that was called to skip the test.
func (t *SatisfyT) skip(fn string) {
	t.skipped = true
	t.explain(fn)
	panic(abortSentinel{})
}

// fail marks the test as failed.
// fn is the name of the function that was called to indicate failure.
func (t *SatisfyT) fail(fn string) {
	t.failed = true
	t.explain(fn)
}

// explain populates t.explanation, including file/line information.
func (t *SatisfyT) explain(fn string) {
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

// close calls any functions registered via Cleanup().
func (t *SatisfyT) close() {
	for i := len(t.cleanup) - 1; i >= 0; i-- {
		t.cleanup[i]()
	}
}

// findFrame searches, starting after skip frames, for the first caller frame
// in a function not marked as a helper and returns that frame.
//
// The search stops at the frame where the user-supplied expectation function is called.
//
// It is assumed that s.m is already locked.
func (t *SatisfyT) findFrame(skip int) (runtime.Frame, bool) {
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

// stack returns frames *above the caller* on the stack.
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
