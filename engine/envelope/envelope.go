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
	// This value is nil unless Role is message.TimeoutRole.
	TimeoutTime *time.Time
}

// New constructs a new envelope containing the given message.
func New(m dogma.Message, r message.Role) *Envelope {
	r.MustValidate()

	return &Envelope{
		Message: m,
		Type:    message.TypeOf(m),
		Role:    r,
		IsRoot:  true,
	}
}

// NewChild constructs a new envelope as a child of e, indicating that m is
// caused by e.Message.
func (e *Envelope) NewChild(m dogma.Message, r message.Role) *Envelope {
	r.MustValidate()

	return &Envelope{
		Message: m,
		Type:    message.TypeOf(m),
		Role:    r,
		IsRoot:  false,
	}
}
