package dogmatest

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/assert"
	"github.com/dogmatiq/enginekit/message"
)

// ExpectEvent returns an assertion the requires m to be published as an event.
func ExpectEvent(m dogma.Message) assert.Assertion {
	return &assert.MessageAssertion{
		Message: m,
		Role:    message.EventRole,
	}
}
