package testkit

import (
	"context"
	"sync"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine"
)

// CommandExecutor is an implementation of dogma.CommandExecutor that executes
// commands within the context of a Test.
//
// Each instance is bound to a particular Test. Use Test.CommandExecutor() to
// obtain an instance.
type CommandExecutor struct {
	m           sync.RWMutex
	next        engine.CommandExecutor
	interceptor CommandExecutorInterceptor
}

// ExecuteCommand executes the command message m.
//
// It panics unless it is called during an Action, such as when calling
// Test.Prepare() or Test.Expect().
func (c *CommandExecutor) ExecuteCommand(ctx context.Context, m dogma.Message) error {
	c.m.RLock()
	defer c.m.RUnlock()

	if c.next.Engine == nil {
		panic("ExecuteCommand(): can not be called outside of a test")
	}

	if c.interceptor != nil {
		return c.interceptor(ctx, m, c.next)
	}

	return c.next.ExecuteCommand(ctx, m)
}

// Bind sets the engine and options used to execute commands.
//
// It is intended for use within Action implementations that support executing
// commands outside of a Dogma handler, such as Call().
//
// It must be called before ExecuteCommand(), otherwise ExecuteCommand() panics.
//
// It must be accompanied by a call to Unbind() upon completion of the Action.
func (c *CommandExecutor) Bind(e *engine.Engine, options []engine.OperationOption) {
	c.m.Lock()
	defer c.m.Unlock()

	c.next.Engine = e
	c.next.Options = options
}

// Unbind removes the engine and options configured by a prior call to Bind().
//
// Calls to ExecuteCommand() on an unbound executor will cause a panic.
func (c *CommandExecutor) Unbind() {
	c.m.Lock()
	defer c.m.Unlock()

	c.next.Engine = nil
	c.next.Options = nil
}

// Intercept installs an interceptor function that is invoked whenever
// ExecuteCommand() is called.
//
// If fn is nil the interceptor is removed.
//
// It returns the previous interceptor, if any.
func (c *CommandExecutor) Intercept(fn CommandExecutorInterceptor) CommandExecutorInterceptor {
	c.m.Lock()
	defer c.m.Unlock()

	prev := c.interceptor
	c.interceptor = fn

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
	m dogma.Message,
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
	a.onExec = o.fn
}
