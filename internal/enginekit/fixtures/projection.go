package fixtures

import (
	"context"

	"github.com/dogmatiq/dogma"
)

// ProjectionMessageHandler is a test implementation of dogma.ProjectionMessageHandler.
type ProjectionMessageHandler struct {
	ConfigureFunc   func(dogma.ProjectionConfigurer)
	HandleEventFunc func(context.Context, dogma.ProjectionEventScope, dogma.Message) error
}

var _ dogma.ProjectionMessageHandler = &ProjectionMessageHandler{}

// Configure configures the behavior of the engine as it relates to this
// handler.
//
// c provides access to the various configuration options, such as specifying
// which types of event messages are routed to this handler.
//
// If h.ConfigureFunc is non-nil, it calls h.ConfigureFunc(c)
func (h *ProjectionMessageHandler) Configure(c dogma.ProjectionConfigurer) {
	if h.ConfigureFunc != nil {
		h.ConfigureFunc(c)
	}
}

// HandleEvent handles a domain event message that has been routed to this
// handler.
//
// s provides access to the operations available within the scope of handling
// m.
//
// It panics with the UnexpectedMessage value if m is not one of the event
// types that is routed to this handler via Configure().
//
// If h.HandleEventFunc is non-nil it calls h.HandleEventFunc(s, m).
func (h *ProjectionMessageHandler) HandleEvent(
	ctx context.Context,
	s dogma.ProjectionEventScope,
	m dogma.Message,
) error {
	if h.HandleEventFunc != nil {
		return h.HandleEventFunc(ctx, s, m)
	}

	return nil
}
