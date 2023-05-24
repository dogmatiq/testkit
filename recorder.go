package testkit

import (
	"context"
	"sync"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine"
)

// EventRecorder is an implementation of dogma.EventRecorder that records events
// within the context of a Test.
//
// Each instance is bound to a particular Test. Use Test.EventRecorder() to
// obtain an instance.
type EventRecorder struct {
	m           sync.RWMutex
	next        engine.EventRecorder
	interceptor EventRecorderInterceptor
}

// RecordEvent records the event message m.
//
// It panics unless it is called during an Action, such as when calling
// Test.Prepare() or Test.Expect().
func (r *EventRecorder) RecordEvent(ctx context.Context, m dogma.Message) error {
	r.m.RLock()
	defer r.m.RUnlock()

	if r.next.Engine == nil {
		panic("RecordEvent(): cannot be called outside of a test")
	}

	if r.interceptor != nil {
		return r.interceptor(ctx, m, r.next)
	}

	return r.next.RecordEvent(ctx, m)
}

// Bind sets the engine and options used to record events.
//
// It is intended for use within Action implementations that support recording
// events outside of a Dogma handler, such as Call().
//
// It must be called before RecordEvent(), otherwise RecordEvent() panics.
//
// It must be accompanied by a call to Unbind() upon completion of the Action.
func (r *EventRecorder) Bind(eng *engine.Engine, options []engine.OperationOption) {
	r.m.Lock()
	defer r.m.Unlock()

	r.next.Engine = eng
	r.next.Options = options
}

// Unbind removes the engine and options configured by a prior call to Bind().
//
// Calls to RecordEvent() on an unbound recorder will cause a panic.
func (r *EventRecorder) Unbind() {
	r.m.Lock()
	defer r.m.Unlock()

	r.next.Engine = nil
	r.next.Options = nil
}

// Intercept installs an interceptor function that is invoked whenever
// RecordEvent() is called.
//
// If fn is nil the interceptor is removed.
//
// It returns the previous interceptor, if any.
func (r *EventRecorder) Intercept(fn EventRecorderInterceptor) EventRecorderInterceptor {
	r.m.Lock()
	defer r.m.Unlock()

	prev := r.interceptor
	r.interceptor = fn

	return prev
}

// EventRecorderInterceptor is used by the InterceptEventRecorder() option to
// specify custom behavior for the dogma.EventRecorder returned by
// Test.EventRecorder().
//
// m is the event being recorded.
//
// e can be used to record the event as it would be recorded without this
// interceptor installed.
type EventRecorderInterceptor func(
	ctx context.Context,
	m dogma.Message,
	r dogma.EventRecorder,
) error

// InterceptEventRecorder returns an option that causes fn to be called
// whenever an event is recorded via the dogma.EventRecorder returned by
// Test.EventRecorder().
//
// Intercepting calls to the event recorder allows the user to simulate
// failures (or any other behavior) in the event recorder.
func InterceptEventRecorder(fn EventRecorderInterceptor) interface {
	TestOption
	CallOption
} {
	if fn == nil {
		panic("InterceptEventRecorder(<nil>): function must not be nil")
	}

	return interceptEventRecorderOption{fn}
}

// interceptEventRecorderOption is an implementation of both TestOption and
// CallOption that allows the InterceptEventRecorder() option to be used with
// both Test.Begin() and Call().
type interceptEventRecorderOption struct {
	fn EventRecorderInterceptor
}

func (o interceptEventRecorderOption) applyTestOption(t *Test) {
	t.recorder.Intercept(o.fn)
}

func (o interceptEventRecorderOption) applyCallOption(a *callAction) {
	a.onRecord = o.fn
}
