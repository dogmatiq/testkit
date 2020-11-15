package testkit

import (
	"context"
	"fmt"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/internal/inflect"
)

// ExecuteCommand returns an Action that executes a command message.
func ExecuteCommand(m dogma.Message) Action {
	return executeCommand{m}
}

// executeCommand is an implementation of Action that executes a command
// message.
type executeCommand struct {
	m dogma.Message
}

// Heading returns a human-readable description of the action, used as a
// heading within the test report.
//
// Any engine activity as a result of this action is logged beneath this
// heading.
func (a executeCommand) Heading() string {
	return fmt.Sprintf("EXECUTING TEST COMMAND (%T)", a.m)
}

// ExpectOptions returns the options to use by default when this action is
// used with Test.Expect().
func (a executeCommand) ExpectOptions() []ExpectOption {
	return nil
}

// Apply performs the action within the context of a specific test.
func (a executeCommand) Apply(ctx context.Context, s ActionScope) error {
	mt := message.TypeOf(a.m)
	r, ok := s.App.MessageTypes().RoleOf(mt)

	if !ok {
		return inflect.Errorf(
			r,
			"can not execute %T as a command, it is a not a recognized message type",
			a.m,
		)
	} else if r != message.CommandRole {
		return inflect.Errorf(
			r,
			"can not execute %T as a command, it is configured as a <message>",
			a.m,
		)
	}

	return s.Engine.Dispatch(ctx, a.m, s.OperationOptions...)
}
