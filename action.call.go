package testkit

import (
	"context"

	"github.com/dogmatiq/testkit/location"
)

// Call is an Action that invokes a user-defined function within the context of
// a test.
//
// It is intended to execute application code that makes use of the
// dogma.CommandExecutor or dogma.EventRecorder interfaces.
//
// Test implementations of these interfaces can be OBTAINED via the
// Test.CommandExecutor() and Test.EventRecorder() methods at any time; however,
// they may only be USED within a function invoked by a Call() action.
//
// When Call() is used with Test.Expect() the expectation will match the
// messages dispatched via the test's executor and recorder, as well as those
// produced by handlers within the Dogma application.
func Call(fn func(), options ...CallOption) Action {
	if fn == nil {
		panic("Call(<nil>): function must not be nil")
	}

	act := callAction{
		fn:  fn,
		loc: location.OfCall(),
	}

	for _, opt := range options {
		opt.applyCallOption(&act)
	}

	return act
}

// CallOption applies optional settings to a Call action.
type CallOption interface {
	applyCallOption(*callAction)
}

// callAction is an implementation of Action that invokes a user-defined
// function.
type callAction struct {
	fn        func()
	loc       location.Location
	onExecute CommandExecutorInterceptor
	onRecord  EventRecorderInterceptor
}

func (a callAction) Caption() string {
	return "calling user-defined function"
}

func (a callAction) Location() location.Location {
	return a.loc
}

func (a callAction) ConfigurePredicate(o *PredicateOptions) {
	o.MatchDispatchCycleStartedFacts = true
}

func (a callAction) Do(ctx context.Context, s ActionScope) error {
	// Setup the command executor for use during this action.
	s.Executor.Bind(s.Engine, s.OperationOptions)
	defer s.Executor.Unbind()

	if a.onExecute != nil {
		prev := s.Executor.Intercept(a.onExecute)
		defer s.Executor.Intercept(prev)
	}

	// Setup the event recorder for use during this action.
	s.Recorder.Bind(s.Engine, s.OperationOptions)
	defer s.Recorder.Unbind()

	if a.onRecord != nil {
		prev := s.Recorder.Intercept(a.onRecord)
		defer s.Recorder.Intercept(prev)
	}

	// Execute the user-supplied function.
	a.fn()

	return nil
}
