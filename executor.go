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
// An instance can be obtained at any time by calling Test.CommandExecutor().
type commandExecutor struct {
	m    sync.RWMutex
	next engine.CommandExecutor
}

// ExecuteCommand executes the command message m.
func (c *commandExecutor) ExecuteCommand(ctx context.Context, m dogma.Message) error {
	c.m.RLock()
	defer c.m.RUnlock()

	if c.next.Engine != nil {
		return c.next.ExecuteCommand(ctx, m)
	}

	panic("ExecuteCommand(): can not be called outside of a test action")
}

// bind sets the engine and options that should be used to execute commands.
//
// Binding to an engine allowed the executor to be used. A call to unbind() must
// be made when the executor should not longer be used.
func (c *commandExecutor) bind(e *engine.Engine, options []engine.OperationOption) {
	c.m.Lock()
	defer c.m.Unlock()

	c.next.Engine = e
	c.next.Options = options
}

// unbind removes the engine and options configured by the prior call to bind().
//
// Calling ExecuteCommand() on an unbound executor causes a panic.
func (c *commandExecutor) unbind() {
	c.m.Lock()
	defer c.m.Unlock()

	c.next.Engine = nil
	c.next.Options = nil
}
