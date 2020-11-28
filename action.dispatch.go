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
	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf("ExecuteCommand(%T): %s", m, err))
	}

	return dispatchAction{message.CommandRole, m}
}

// RecordEvent returns an Action that records an event message.
func RecordEvent(m dogma.Message) Action {
	if err := dogma.ValidateMessage(m); err != nil {
		panic(fmt.Sprintf("RecordEvent(%T): %s", m, err))
	}

	return dispatchAction{message.EventRole, m}
}

// dispatchAction is an implementation of Action that dispatches a message to
// the engine.
type dispatchAction struct {
	r message.Role
	m dogma.Message
}

func (a dispatchAction) Banner() string {
	return inflect.Sprintf(
		a.r,
		"<PRODUCING> %T <MESSAGE>",
		a.m,
	)
}

func (a dispatchAction) ConfigurePredicate(o *PredicateOptions) {
}

func (a dispatchAction) Do(ctx context.Context, s ActionScope) error {
	mt := message.TypeOf(a.m)
	r, ok := s.App.MessageTypes().RoleOf(mt)

	// TODO: These checks should result in information being added to the
	// report, not just returning an error.
	//
	// See https://github.com/dogmatiq/testkit/issues/162
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
