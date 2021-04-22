package testkit

import (
	"context"
	"sync"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/engine"
)

// commandExecutor is an implementation of dogma.CommandExecutor that executes
// commands within the context of a test.
//
// An instance can be obtained at any time by calling Test.CommandExecutor(),
// but the executor can only be used within an Action.
type commandExecutor struct {
	m           sync.RWMutex
	next        engine.CommandExecutor
	interceptor CommandExecutorInterceptor
}

// ExecuteCommand executes the command message m.
//
// It panics unless an Action is in progress. See bind() and unbind().
func (c *commandExecutor) ExecuteCommand(ctx context.Context, m dogma.Message) error {
	c.m.RLock()
	defer c.m.RUnlock()

	if c.next.Engine == nil {
		panic("ExecuteCommand(): can not be called outside of a test action")
	}

	if c.interceptor != nil {
		return c.interceptor(ctx, m, c.next)
	}

	return c.next.ExecuteCommand(ctx, m)
}

// bind sets the engine and options used to execute commands.
//
// It is called at the start of each Action with options that allow the Test to
// inspect the facts produced by the commands executed via this executor.
//
// unbind() must be called after each action.
func (c *commandExecutor) bind(e *engine.Engine, options []engine.OperationOption) {
	c.m.Lock()
	defer c.m.Unlock()

	c.next.Engine = e
	c.next.Options = options
}

// unbind removes the engine and options configured by a prior call to bind().
func (c *commandExecutor) unbind() {
	c.m.Lock()
	defer c.m.Unlock()

	c.next.Engine = nil
	c.next.Options = nil
}

// intercept installs an interceptor function that is invoked whenever
// ExecuteCommand() is called.
//
// If fn is nil the interceptor is removed.
//
// It returns the previous interceptor, if any.
func (c *commandExecutor) intercept(fn CommandExecutorInterceptor) CommandExecutorInterceptor {
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
func InterceptCommandExecutor(fn CommandExecutorInterceptor) TestOption {
	return testOptionFunc(func(t *Test) {
		t.executor.intercept(fn)
	})
}
