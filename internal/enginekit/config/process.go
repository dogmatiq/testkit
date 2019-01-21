package config

import (
	"context"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// ProcessConfig represents the configuration of an process message handler.
type ProcessConfig struct {
	// Handler is the handler that the configuration applies to.
	Handler dogma.ProcessMessageHandler

	// HandlerName is the handler's name, as specified by its Configure() method.
	HandlerName string

	// MessageTypes is the set of event message types that are routed to this
	// handler, as specified by its Configure() method.
	MessageTypes message.TypeSet
}

// NewProcessConfig returns an ProcessConfig for the given handler.
func NewProcessConfig(h dogma.ProcessMessageHandler) (*ProcessConfig, error) {
	cfg := &ProcessConfig{
		Handler:      h,
		MessageTypes: message.TypeSet{},
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

	if len(c.cfg.MessageTypes) == 0 {
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

// HandlerType returns handler.ProcessType.
func (c *ProcessConfig) HandlerType() handler.Type {
	return handler.ProcessType
}

// HandlerReflectType returns the reflect.Type of the handler.
func (c *ProcessConfig) HandlerReflectType() reflect.Type {
	return reflect.TypeOf(c.Handler)
}

// CommandTypes returns the types of command messages that are routed to the handler.
func (c *ProcessConfig) CommandTypes() message.TypeSet {
	return nil
}

// EventTypes returns the types of event messages that are routed to the handler.
func (c *ProcessConfig) EventTypes() message.TypeSet {
	return c.MessageTypes
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

	if _, ok := c.cfg.MessageTypes[t]; ok {
		panicf(
			`%T.Configure() has already called ProcessConfigurer.RouteEventType(%T)`,
			c.cfg.Handler,
			m,
		)
	}

	c.cfg.MessageTypes[t] = struct{}{}
}
