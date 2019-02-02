package envelope

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
)

// Envelope is a container for a message that is handled by the test engine.
type Envelope struct {
	message.Correlation

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
	id string,
	m dogma.Message,
	r message.Role,
) *Envelope {
	if id == "" {
		panic("message ID must not be empty")
	}

	r.MustNotBe(message.TimeoutRole)

	c := message.NewCorrelation(id)
	c.MustValidate()

	return &Envelope{
		Correlation: c,
		Message:     m,
		Type:        message.TypeOf(m),
		Role:        r,
	}
}

// NewCommand constructs a new command envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewCommand(
	id string,
	m dogma.Message,
	o Origin,
) *Envelope {
	return e.new(id, m, message.CommandRole, o)
}

// NewEvent constructs a new event envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewEvent(
	id string,
	m dogma.Message,
	o Origin,
) *Envelope {
	return e.new(id, m, message.EventRole, o)
}

// NewTimeout constructs a new event envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewTimeout(
	id string,
	m dogma.Message,
	t time.Time,
	o Origin,
) *Envelope {
	env := e.new(id, m, message.TimeoutRole, o)
	env.TimeoutTime = &t
	return env
}

func (e *Envelope) new(
	id string,
	m dogma.Message,
	r message.Role,
	o Origin,
) *Envelope {
	c := e.Correlation.New(id)
	c.MustValidate()

	return &Envelope{
		Correlation: c,
		Message:     m,
		Type:        message.TypeOf(m),
		Role:        r,
		Origin:      &o,
	}
}
