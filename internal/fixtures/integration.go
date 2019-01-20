package fixtures

import (
	"context"

	"github.com/dogmatiq/dogma"
)

// IntegrationMessageHandler is a test implementation of dogma.IntegrationMessageHandler.
type IntegrationMessageHandler struct {
	ConfigureFunc     func(dogma.IntegrationConfigurer)
	HandleCommandFunc func(context.Context, dogma.IntegrationCommandScope, dogma.Message) error
}

var _ dogma.IntegrationMessageHandler = &IntegrationMessageHandler{}

// Configure configures the behavior of the engine as it relates to this
// handler.
//
// c provides access to the various configuration options, such as specifying
// which types of integration command messages are routed to this handler.
//
// If h.ConfigureFunc is non-nil, it calls h.ConfigureFunc(c)
func (h *IntegrationMessageHandler) Configure(c dogma.IntegrationConfigurer) {
	if h.ConfigureFunc != nil {
		h.ConfigureFunc(c)
	}
}

// HandleCommand handles an integration command message that has been routed to
// this handler.
//
// s provides access to the operations available within the scope of handling
// m, such as publishing integration event messages.
//
// It panics with the UnexpectedMessage value if m is not one of the
// integration command types that is routed to this handler via Configure().
//
// If h.HandleCommandFunc is non-nil it calls h.HandleCommandFunc(s, m),
// otherwise it panics.
func (h *IntegrationMessageHandler) HandleCommand(
	ctx context.Context,
	s dogma.IntegrationCommandScope,
	m dogma.Message,
) error {
	if h.HandleCommandFunc == nil {
		panic(dogma.UnexpectedMessage)
	}

	return h.HandleCommandFunc(ctx, s, m)
}
