package envelope

import "github.com/dogmatiq/enginekit/config"

// Origin describes the handler that produced a message in an envelope.
type Origin struct {
	// Handler is the handler that produced this message.
	Handler config.Handler

	// InstanceID is the ID of the aggregate or process instance that produced
	// this message.
	InstanceID string
}
