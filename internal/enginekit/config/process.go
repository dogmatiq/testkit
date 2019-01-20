package config

import (
	"context"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// ProcessConfig represents the configuration of an process message handler.
type ProcessConfig struct {
	// Handler is the handler that the configuration applies to.
	Handler dogma.ProcessMessageHandler

	// HandlerName is the handler's name, as specified by its Configure() method.
	HandlerName string

	// EventTypes is the set of event message types that are routed to this
	// handler, as specified by its Configure() method.
	EventTypes map[message.Type]struct{}
}

// NewProcessConfig returns an ProcessConfig for the given handler.
func NewProcessConfig(h dogma.ProcessMessageHandler) (*ProcessConfig, error) {
	cfg := &ProcessConfig{
		Handler:    h,
		EventTypes: map[message.Type]struct{}{},
	}

	c := &processConfigurer{
		cfg: cfg,
	}

	if err := catch(func() {
		h.Configure(c)
	}); err != nil {
		return nil, err
	}

	if c.cfg.HandlerName == "" {
		return nil, errorf(
			"%T.Configure() did not call ProcessConfigurer.Name()",
			h,
		)
	}

	if len(c.cfg.EventTypes) == 0 {
		return nil, errorf(
			"%T.Configure() did not call ProcessConfigurer.RouteEventType()",
			h,
		)
	}

	return cfg, nil
}

// Name returns the process name.
func (c *ProcessConfig) Name() string {
	return c.HandlerName
}

// Accept calls v.VisitProcessConfig(ctx, c).
func (c *ProcessConfig) Accept(ctx context.Context, v Visitor) error {
	return v.VisitProcessConfig(ctx, c)
}

// processConfigurer is an implementation of dogma.ProcessConfigurer
// that builds an ProcessConfig value.
type processConfigurer struct {
	cfg *ProcessConfig
}

func (c *processConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panicf(
			`%T.Configure() has already called ProcessConfigurer.Name(%#v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		)
	}

	if strings.TrimSpace(n) == "" {
		panicf(
			`%T.Configure() called ProcessConfigurer.Name(%#v) with an invalid name`,
			c.cfg.Handler,
			n,
		)
	}

	c.cfg.HandlerName = n
}

func (c *processConfigurer) RouteEventType(m dogma.Message) {
	t := message.TypeOf(m)

	if _, ok := c.cfg.EventTypes[t]; ok {
		panicf(
			`%T.Configure() has already called ProcessConfigurer.RouteEventType(%T)`,
			c.cfg.Handler,
			m,
		)
	}

	c.cfg.EventTypes[t] = struct{}{}
}
