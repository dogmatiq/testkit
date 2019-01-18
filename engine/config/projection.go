package config

import (
	"context"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
)

// ProjectionConfig represents the configuration of an aggregate message handler.
type ProjectionConfig struct {
	// Handler is the handler that the configuration applies to.
	Handler dogma.ProjectionMessageHandler

	// HandlerName is the handler's name, as specified by its Configure() method.
	HandlerName string

	// EventTypes is the set of event message types that are routed to this
	// handler, as specified by its Configure() method.
	EventTypes map[reflect.Type]struct{}
}

// NewProjectionConfig returns an ProjectionConfig for the given handler.
func NewProjectionConfig(h dogma.ProjectionMessageHandler) (*ProjectionConfig, error) {
	cfg := &ProjectionConfig{
		Handler:    h,
		EventTypes: map[reflect.Type]struct{}{},
	}

	c := &projectionConfigurer{
		cfg: cfg,
	}

	if err := catch(func() {
		h.Configure(c)
	}); err != nil {
		return nil, err
	}

	if c.cfg.HandlerName == "" {
		return nil, errorf(
			"%T.Configure() did not call ProjectionConfigurer.Name()",
			h,
		)
	}

	if len(c.cfg.EventTypes) == 0 {
		return nil, errorf(
			"%T.Configure() did not call ProjectionConfigurer.RouteEventType()",
			h,
		)
	}

	return cfg, nil
}

// Name returns the projection name.
func (c *ProjectionConfig) Name() string {
	return c.HandlerName
}

// Accept calls v.VisitProjectionConfig(ctx, c).
func (c *ProjectionConfig) Accept(ctx context.Context, v Visitor) error {
	return v.VisitProjectionConfig(ctx, c)
}

// projectionConfigurer is an implementation of dogma.ProjectionConfigurer
// that builds an ProjectionConfig value.
type projectionConfigurer struct {
	cfg *ProjectionConfig
}

func (c *projectionConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panicf(
			`%T.Configure() has already called ProjectionConfigurer.Name(%#v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		)
	}

	if strings.TrimSpace(n) == "" {
		panicf(
			`%T.Configure() called ProjectionConfigurer.Name(%#v) with an invalid name`,
			c.cfg.Handler,
			n,
		)
	}

	c.cfg.HandlerName = n
}

func (c *projectionConfigurer) RouteEventType(m dogma.Message) {
	t := reflect.TypeOf(m)

	if _, ok := c.cfg.EventTypes[t]; ok {
		panicf(
			`%T.Configure() has already called ProjectionConfigurer.RouteEventType(%T)`,
			c.cfg.Handler,
			m,
		)
	}

	c.cfg.EventTypes[t] = struct{}{}
}
