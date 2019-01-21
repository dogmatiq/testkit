package envelope

import (
	"time"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// Envelope is a container for a message that is handled by the test engine.
type Envelope struct {
	// Message is the application-defined message that the envelope represents.
	Message dogma.Message

	// Type is the type of the message.
	Type message.Type

	// Role is the message's role.
	Role message.Role

	// IsRoot is true if this message is at the root of an envelope tree.
	IsRoot bool

	// TimeoutTime holds the time at which the timeout is scheduled.
	// If the Role is not message.TimeoutRole, this value is nil.
	TimeoutTime *time.Time
}

// New constructs a new envelope containing the given message.
func New(m dogma.Message, r message.Role) *Envelope {
	r.MustValidate()

	if r == message.TimeoutRole {
		panic("the root message can not be a timeout")
	}

	return &Envelope{
		Message: m,
		Type:    message.TypeOf(m),
		Role:    r,
		IsRoot:  true,
	}
}

// NewCommand constructs a new command envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewCommand(m dogma.Message) *Envelope {
	return &Envelope{
		Message: m,
		Type:    message.TypeOf(m),
		Role:    message.CommandRole,
		IsRoot:  false,
	}
}

// NewEvent constructs a new event envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewEvent(m dogma.Message) *Envelope {
	return &Envelope{
		Message: m,
		Type:    message.TypeOf(m),
		Role:    message.EventRole,
		IsRoot:  false,
	}
}

// NewTimeout constructs a new event envelope as a child of e, indicating that
// m is caused by e.Message.
func (e *Envelope) NewTimeout(m dogma.Message, t time.Time) *Envelope {
	return &Envelope{
		Message:     m,
		Type:        message.TypeOf(m),
		Role:        message.TimeoutRole,
		IsRoot:      false,
		TimeoutTime: &t,
	}
}
