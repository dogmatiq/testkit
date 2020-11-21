package testkit

import (
	"context"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/internal/inflect"
)

// ExecuteCommand returns an Action that executes a command message.
func ExecuteCommand(m dogma.Message) Action {
	if m == nil {
		panic("ExecuteCommand(): message must not be nil")
	}

	return dispatch{message.CommandRole, m}
}

// RecordEvent returns an Action that records an event message.
func RecordEvent(m dogma.Message) Action {
	if m == nil {
		panic("RecordEvent(): message must not be nil")
	}

	return dispatch{message.EventRole, m}
}

// dispatchMessage is an implementation of Action that dispatches a message to
// the engine.
type dispatch struct {
	r message.Role
	m dogma.Message
}

// Banner returns a human-readable banner to display in the logs when this
// action is applied.
//
// The banner text should be in uppercase, and worded in the present tense,
// for example "DOING ACTION".
func (a dispatch) Banner() string {
	return inflect.Sprintf(
		a.r,
		"<PRODUCING> %T <MESSAGE>",
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
