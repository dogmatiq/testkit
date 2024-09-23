package testkit

import (
	"context"
	"fmt"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/testkit/internal/inflect"
	"github.com/dogmatiq/testkit/location"
)

// ExecuteCommand returns an Action that executes a command message.
func ExecuteCommand(m dogma.Command) Action {
	if m == nil {
		panic("ExecuteCommand(<nil>): message must not be nil")
	}

	mt := message.TypeOf(m)

	if err := m.Validate(); err != nil {
		panic(fmt.Sprintf("ToRecordEvent(%s): %s", mt, err))
	}

	return dispatchAction{
		message.CommandRole,
		m,
		location.OfCall(),
	}
}

// RecordEvent returns an Action that records an event message.
func RecordEvent(m dogma.Event) Action {
	if m == nil {
		panic("RecordEvent(<nil>): message must not be nil")
	}

	mt := message.TypeOf(m)

	if err := m.Validate(); err != nil {
		panic(fmt.Sprintf("RecordEvent(%s): %s", mt, err))
	}

	return dispatchAction{
		message.EventRole,
		m,
		location.OfCall(),
	}
}

// dispatchAction is an implementation of Action that dispatches a message to
// the engine.
type dispatchAction struct {
	r   message.Role
	m   dogma.Message
	loc location.Location
}

func (a dispatchAction) Caption() string {
	return inflect.Sprintf(
		a.r,
		"<producing> %s <message>",
		message.TypeOf(a.m),
	)
}

func (a dispatchAction) Location() location.Location {
	return a.loc
}

func (a dispatchAction) ConfigurePredicate(*PredicateOptions) {
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
			"cannot <produce> <message>, %s is a not a recognized message type",
			mt,
		)
	} else if r != a.r {
		return inflect.Errorf(
			a.r,
			"cannot <produce> <message>, %s",
			inflect.Sprintf(
				r,
				"%s is configured as a <message>",
				mt,
			),
		)
	}

	return s.Engine.Dispatch(ctx, a.m, s.OperationOptions...)
}
