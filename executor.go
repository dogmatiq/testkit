package testkit

import (
	"context"
	"sync"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine"
)

// CommandExecutor is an implementation of [dogma.CommandExecutor] that executes
// commands within the context of a Test.
//
// Each instance is bound to a particular Test. Use Test.CommandExecutor() to
// obtain an instance.
type CommandExecutor struct {
	m           sync.RWMutex
	next        engine.CommandExecutor
	interceptor CommandExecutorInterceptor
}

var _ dogma.CommandExecutor = (*CommandExecutor)(nil)

// ExecuteCommand executes the command message m.
//
// It panics unless it is called during an Action, such as when calling
// Test.Prepare() or Test.Expect().
func (e *CommandExecutor) ExecuteCommand(
	ctx context.Context,
	m dogma.Command,
	_ ...dogma.ExecuteCommandOption,
) error {
	e.m.RLock()
	defer e.m.RUnlock()

	if e.next.Engine == nil {
		panic("ExecuteCommand(): cannot be called outside of a test")
	}

	if e.interceptor != nil {
		return e.interceptor(ctx, m, e.next)
	}

	return e.next.ExecuteCommand(ctx, m)
}

// Bind sets the engine and options used to execute commands.
//
// It is intended for use within Action implementations that support executing
// commands outside of a Dogma handler, such as Call().
//
// It must be called before ExecuteCommand(), otherwise ExecuteCommand() panics.
//
// It must be accompanied by a call to Unbind() upon completion of the Action.
func (e *CommandExecutor) Bind(eng *engine.Engine, options []engine.OperationOption) {
	e.m.Lock()
	defer e.m.Unlock()

	e.next.Engine = eng
	e.next.Options = options
}

// Unbind removes the engine and options configured by a prior call to Bind().
//
// Calls to ExecuteCommand() on an unbound executor will cause a panic.
func (e *CommandExecutor) Unbind() {
	e.m.Lock()
	defer e.m.Unlock()

	e.next.Engine = nil
	e.next.Options = nil
}

// Intercept installs an interceptor function that is invoked whenever
// ExecuteCommand() is called.
//
// If fn is nil the interceptor is removed.
//
// It returns the previous interceptor, if any.
func (e *CommandExecutor) Intercept(fn CommandExecutorInterceptor) CommandExecutorInterceptor {
	e.m.Lock()
	defer e.m.Unlock()

	prev := e.interceptor
	e.interceptor = fn

	return prev
}

// CommandExecutorInterceptor is used by the InterceptCommandExecutor() option
// to specify custom behavior for the dogma.CommandExecutor returned by
// Test.CommandExecutor().
//
// m is the command being executed.
//
// e can be used to execute the command as it would be executed without this
// interceptor installed.
type CommandExecutorInterceptor func(
	ctx context.Context,
	m dogma.Command,
	e dogma.CommandExecutor,
) error

// InterceptCommandExecutor returns an option that causes fn to be called
// whenever a command is executed via the dogma.CommandExecutor returned by
// Test.CommandExecutor().
//
// Intercepting calls to the command executor allows the user to simulate
// failures (or any other behavior) in the command executor.
func InterceptCommandExecutor(fn CommandExecutorInterceptor) interface {
	TestOption
	CallOption
} {
	if fn == nil {
		panic("InterceptCommandExecutor(<nil>): function must not be nil")
	}

	return interceptCommandExecutorOption{fn}
}

// interceptCommandExecutorOption is an implementation of both TestOption and
// CallOption that allows the InterceptCommandExecutor() option to be used with
// both Test.Begin() and Call().
type interceptCommandExecutorOption struct {
	fn CommandExecutorInterceptor
}

func (o interceptCommandExecutorOption) applyTestOption(t *Test) {
	t.executor.Intercept(o.fn)
}

func (o interceptCommandExecutorOption) applyCallOption(a *callAction) {
	a.onExecute = o.fn
}
