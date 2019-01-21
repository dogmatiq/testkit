package config

import (
	"context"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/internal/enginekit/handler"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// IntegrationConfig represents the configuration of an integration message handler.
type IntegrationConfig struct {
	// Handler is the handler that the configuration applies to.
	Handler dogma.IntegrationMessageHandler

	// HandlerName is the handler's name, as specified by its Configure() method.
	HandlerName string

	// MessageTypes is the set of command message types that are routed to this
	// handler, as specified by its Configure() method.
	MessageTypes message.TypeSet
}

// NewIntegrationConfig returns an IntegrationConfig for the given handler.
func NewIntegrationConfig(h dogma.IntegrationMessageHandler) (*IntegrationConfig, error) {
	cfg := &IntegrationConfig{
		Handler:      h,
		MessageTypes: message.TypeSet{},
	}

	c := &integrationConfigurer{
		cfg: cfg,
	}

	if err := catch(func() {
		h.Configure(c)
	}); err != nil {
		return nil, err
	}

	if c.cfg.HandlerName == "" {
		return nil, errorf(
			"%T.Configure() did not call IntegrationConfigurer.Name()",
			h,
		)
	}

	if len(c.cfg.MessageTypes) == 0 {
		return nil, errorf(
			"%T.Configure() did not call IntegrationConfigurer.RouteCommandType()",
			h,
		)
	}

	return cfg, nil
}

// Name returns the integration name.
func (c *IntegrationConfig) Name() string {
	return c.HandlerName
}

// HandlerType returns handler.IntegrationType.
func (c *IntegrationConfig) HandlerType() handler.Type {
	return handler.IntegrationType
}

// HandlerReflectType returns the reflect.Type of the handler.
func (c *IntegrationConfig) HandlerReflectType() reflect.Type {
	return reflect.TypeOf(c.Handler)
}

// CommandTypes returns the types of command messages that are routed to the handler.
func (c *IntegrationConfig) CommandTypes() message.TypeSet {
	return c.MessageTypes
}

// EventTypes returns the types of event messages that are routed to the handler.
func (c *IntegrationConfig) EventTypes() message.TypeSet {
	return nil
}

// Accept calls v.VisitIntegrationConfig(ctx, c).
func (c *IntegrationConfig) Accept(ctx context.Context, v Visitor) error {
	return v.VisitIntegrationConfig(ctx, c)
}

// integrationConfigurer is an implementation of dogma.integrationConfigurer
// that builds an IntegrationConfig value.
type integrationConfigurer struct {
	cfg *IntegrationConfig
}

func (c *integrationConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panicf(
			`%T.Configure() has already called IntegrationConfigurer.Name(%#v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		)
	}

	if strings.TrimSpace(n) == "" {
		panicf(
			`%T.Configure() called IntegrationConfigurer.Name(%#v) with an invalid name`,
			c.cfg.Handler,
			n,
		)
	}

	c.cfg.HandlerName = n
}

func (c *integrationConfigurer) RouteCommandType(m dogma.Message) {
	t := message.TypeOf(m)

	if _, ok := c.cfg.MessageTypes[t]; ok {
		panicf(
			`%T.Configure() has already called IntegrationConfigurer.RouteCommandType(%T)`,
			c.cfg.Handler,
			m,
		)
	}

	c.cfg.MessageTypes[t] = struct{}{}
}
