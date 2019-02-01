package envelope

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
)

// Envelope is a container for a message that is handled by the test engine.
type Envelope struct {
	// MessageID is a unique identifier for a message.
	MessageID uint64

	// CausationID is the ID of the message that directly caused this message.
	//
	// If this message was not caused by some other message, CausationID is set to
	// MessageID.
	CausationID uint64

	// CorrelationID is the ID of the message the beginning of a causality chain.
	//
	// If this message was not caused by some other message, CorrelationID is set
	// to MessageID.
	CorrelationID uint64

	// Message is the application-defined message that the envelope represents.
	Message dogma.Message

	// Type is the type of the message.
	Type message.Type

	// Role is the message's role.
	Role message.Role

	// TimeoutTime holds the time at which a timeout message is scheduled to occur.
	// It is nil unless Role is message.TimeoutRole.
	TimeoutTime *time.Time

	// Origin describes the message handler that produced this message.
	// It is nil if the message was not produced by a handler.
	Origin *Origin
}

// New constructs a new envelope containing the given message.
func New(
	id uint64,
	m dogma.Message,
	r message.Role,
) *Envelope {
	if id == 0 {
		panic("message ID must not be zero")
	}

	r.MustNotBe(message.TimeoutRole)

	return &Envelope{
		MessageID:     id,
		CausationID:   id,
		CorrelationID: id,
		Message:       m,
		Type:          message.TypeOf(m),
		Role:          r,
	}
}

// NewCommand constructs a new command envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewCommand(
	id uint64,
	m dogma.Message,
	o Origin,
) *Envelope {
	return e.new(id, m, message.CommandRole, o)
}

// NewEvent constructs a new event envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewEvent(
	id uint64,
	m dogma.Message,
	o Origin,
) *Envelope {
	return e.new(id, m, message.EventRole, o)
}

// NewTimeout constructs a new event envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewTimeout(
	id uint64,
	m dogma.Message,
	t time.Time,
	o Origin,
) *Envelope {
	env := e.new(id, m, message.TimeoutRole, o)
	env.TimeoutTime = &t
	return env
}

func (e *Envelope) new(
	id uint64,
	m dogma.Message,
	r message.Role,
	o Origin,
) *Envelope {
	if id == 0 {
		panic("message ID must not be zero")
	}

	return &Envelope{
		MessageID:     id,
		CausationID:   e.MessageID,
		CorrelationID: e.CorrelationID,
		Message:       m,
		Type:          message.TypeOf(m),
		Role:          r,
		Origin:        &o,
	}
}
