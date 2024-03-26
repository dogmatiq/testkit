package envelope

import (
	"time"

	"github.com/dogmatiq/configkit/message"
	"github.com/dogmatiq/dogma"
)

// Envelope is a container for a message that is handled by the test engine.
type Envelope struct {
	// MessageID is a unique identifier for the message.
	MessageID string

	// CausationID is the ID of the message that was being handled when the
	// message identified by MessageID was produced.
	CausationID string

	// CorrelationID is the ID of the "root" message that entered the
	// application to cause the message identified by MessageID, either directly
	// or indirectly.
	CorrelationID string

	// Message is the application-defined message that the envelope represents.
	Message dogma.Message

	// Type is the type of the message.
	Type message.Type

	// Role is the message's role.
	Role message.Role

	// CreatedAt is the time at which the message was created.
	CreatedAt time.Time

	// ScheduledFor holds the time at which a timeout message is scheduled to
	// occur. Its value is undefined unless Role is message.TimeoutRole.
	ScheduledFor time.Time

	// Origin describes the message handler that produced this message.
	// It is nil if the message was not produced by a handler.
	Origin *Origin
}

// NewCommand constructs a new envelope containing the given command message.
//
// t is the time at which the message was created.
func NewCommand(
	id string,
	m dogma.Message,
	t time.Time,
) *Envelope {
	return newEnvelope(id, m, message.CommandRole, t)
}

// NewEvent constructs a new envelope containing the given event message.
//
// t is the time at which the message was created.
func NewEvent(
	id string,
	m dogma.Message,
	t time.Time,
) *Envelope {
	return newEnvelope(id, m, message.EventRole, t)
}

// newEnvelope constructs a newEnvelope envelope containing the given message.
//
// It panics if r is message.TimeoutRole, as a timeout cannot occur except as a
// result of some other message.
func newEnvelope(
	id string,
	m dogma.Message,
	r message.Role,
	t time.Time,
) *Envelope {
	if id == "" {
		panic("message ID must not be empty")
	}

	r.MustNotBe(message.TimeoutRole)

	return &Envelope{
		MessageID:     id,
		CausationID:   id,
		CorrelationID: id,
		Message:       m,
		Type:          message.TypeOf(m),
		Role:          r,
		CreatedAt:     t,
	}
}

// NewCommand constructs a new envelope as a child of e, indicating that the
// command message m is caused by e.Message.
//
// t is the time at which the message was created.
func (e *Envelope) NewCommand(
	id string,
	m dogma.Message,
	t time.Time,
	o Origin,
) *Envelope {
	return e.new(id, m, message.CommandRole, t, o)
}

// NewEvent constructs a new envelope as a child of e, indicating that the event
// message m is caused by e.Message.
//
// t is the time at which the message was created.
func (e *Envelope) NewEvent(
	id string,
	m dogma.Message,
	t time.Time,
	o Origin,
) *Envelope {
	return e.new(id, m, message.EventRole, t, o)
}

// NewTimeout constructs a new envelope as a child of e, indicating that the
// timeout message m is caused by e.Message.
//
// t is the time at which the message was created. s is the time at which the
// timeout is scheduled to occur.
func (e *Envelope) NewTimeout(
	id string,
	m dogma.Message,
	t time.Time,
	s time.Time,
	o Origin,
) *Envelope {
	env := e.new(id, m, message.TimeoutRole, t, o)
	env.ScheduledFor = s
	return env
}

// new constructs a new envelope as a child of e.
func (e *Envelope) new(
	id string,
	m dogma.Message,
	r message.Role,
	t time.Time,
	o Origin,
) *Envelope {
	return &Envelope{
		MessageID:     id,
		CausationID:   e.MessageID,
		CorrelationID: e.CorrelationID,
		Message:       m,
		Type:          message.TypeOf(m),
		Role:          r,
		CreatedAt:     t,
		Origin:        &o,
	}
}
