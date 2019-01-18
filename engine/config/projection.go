package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
)

// ProjectionConfig represents the configuration of an aggregate message handler.
type ProjectionConfig struct {
	Handler     dogma.ProjectionMessageHandler
	HandlerName string
	EventTypes  map[reflect.Type]struct{}
}

// NewProjectionConfig returns an ProjectionConfig for the given handler.
func NewProjectionConfig(h dogma.ProjectionMessageHandler) *ProjectionConfig {
	cfg := &ProjectionConfig{
		Handler:    h,
		EventTypes: map[reflect.Type]struct{}{},
	}

	c := &projectionConfigurer{
		cfg: cfg,
	}

	h.Configure(c)

	if c.cfg.HandlerName == "" {
		panic(fmt.Sprintf(
			"%T.Configure() did not call ProjectionConfigurer.Name()",
			h,
		))
	}

	if len(c.cfg.EventTypes) == 0 {
		panic(fmt.Sprintf(
			"%T.Configure() did not call ProjectionConfigurer.RouteEventType()",
			h,
		))
	}

	return cfg
}

// Name returns the projection name.
func (c *ProjectionConfig) Name() string {
	return c.HandlerName
}

// Accept calls v.VisitProjectionConfig(c).
func (c *ProjectionConfig) Accept(v Visitor) {
	v.VisitProjectionConfig(c)
}

// projectionConfigurer is an implementation of dogma.ProjectionConfigurer
// that builds an ProjectionConfig value.
type projectionConfigurer struct {
	cfg *ProjectionConfig
}

func (c *projectionConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panic(fmt.Sprintf(
			`%T.Configure() has already called ProjectionConfigurer.Name(#%v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		))
	}

	if strings.TrimSpace(n) == "" {
		panic(fmt.Sprintf(
			`%T.Configure() called ProjectionConfigurer.Name(#%v) with an invalid name`,
			c.cfg.Handler,
			n,
		))
	}

	c.cfg.HandlerName = n
}

func (c *projectionConfigurer) RouteEventType(m dogma.Message) {
	t := reflect.TypeOf(m)

	if _, ok := c.cfg.EventTypes[t]; ok {
		panic(fmt.Sprintf(
			`%T.Configure() has already called ProjectionConfigurer.RouteEventType(%T)`,
			c.cfg.Handler,
			m,
		))
	}

	c.cfg.EventTypes[t] = struct{}{}
}
