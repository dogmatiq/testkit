package testkit

import (
	"context"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/internal/inflect"
)

// ExecuteCommand returns an Action that executes a command message.
func ExecuteCommand(m dogma.Message) Action {
	return dispatch{message.CommandRole, m}
}

// RecordEvent returns an Action that records an event message.
func RecordEvent(m dogma.Message) Action {
	return dispatch{message.EventRole, m}
}

// dispatchMessage is an implementation of Action that dispatches a message to
// the engine.
type dispatch struct {
	r message.Role
	m dogma.Message
}

// Heading returns a human-readable description of the action, used as a
// heading within the test report.
//
// Any engine activity as a result of this action is logged beneath this
// heading.
func (a dispatch) Heading() string {
	return inflect.Sprintf(
		a.r,
		"<PRODUCING> TEST <MESSAGE> (%T)",
		a.m,
	)
}

// ExpectOptions returns the options to use by default when this action is
// used with Test.Expect().
func (a dispatch) ExpectOptions() []ExpectOption {
	return nil
}

// Apply performs the action within the context of a specific test.
func (a dispatch) Apply(ctx context.Context, s ActionScope) error {
	mt := message.TypeOf(a.m)
	r, ok := s.App.MessageTypes().RoleOf(mt)

	if !ok {
		return inflect.Errorf(
			a.r,
			"can not <produce> <message>, %T is a not a recognized message type",
			a.m,
		)
	} else if r != a.r {
		return inflect.Errorf(
			a.r,
			"can not <produce> <message>, %s",
			inflect.Sprintf(
				r,
				"%T is configured as a <message>",
				a.m,
			),
		)
	}

	return s.Engine.Dispatch(ctx, a.m, s.OperationOptions...)
}
