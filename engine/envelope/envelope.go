package envelope

import (
	"reflect"

	"github.com/dogmatiq/dogma"
)

// Envelope is a container for a message that is handled by the test engine.
//
// Envelopes form a tree structure describing the causal relationships between
// the messages they contain.
type Envelope struct {
	// Message is the application-defined message that the envelope represents.
	Message dogma.Message

	// Type is the type of the message.
	Type reflect.Type

	// Role is the message's role.
	Role MessageRole

	// IsHandled is true if the message in the envelope has already been handled.
	IsHandled bool

	// Children is a slice of the envelopes that were caused by this envelope's
	// message when it was handled.
	Children []*Envelope
}

// New constructs a new envelope containing the given message.
func New(m dogma.Message, r MessageRole) *Envelope {
	r.MustValidate()

	return &Envelope{
		Message: m,
		Type:    reflect.TypeOf(m),
		Role:    r,
	}
}

// NewChild constructs a new envelope as a child of e, indicating that m is
// caused by e.Message.
func (e *Envelope) NewChild(m dogma.Message, r MessageRole) *Envelope {
	r.MustValidate()

	env := &Envelope{
		Message: m,
		Type:    reflect.TypeOf(m),
		Role:    r,
	}

	e.Children = append(e.Children, env)

	return env
}

// Walk traverses the "envelope tree", calling fn for each of this envelope's
// children, then their children, recursively.
//
// Traversal is depth-first. fn is never called with e itself, only its
// children.
//
// Traversal stops when fn returns false, or all nodes have been visited.
// It returns true if traversal visits all nodes.
func (e *Envelope) Walk(fn func(*Envelope) bool) bool {
	for _, env := range e.Children {
		if !fn(env) {
			return false
		}

		if !env.Walk(fn) {
			return false
		}
	}

	return true
}
