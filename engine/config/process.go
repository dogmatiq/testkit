package config

import (
	"context"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
)

// ProcessConfig represents the configuration of an process message handler.
type ProcessConfig struct {
	Handler     dogma.ProcessMessageHandler
	HandlerName string
	EventTypes  map[reflect.Type]struct{}
}

// NewProcessConfig returns an ProcessConfig for the given handler.
func NewProcessConfig(h dogma.ProcessMessageHandler) (*ProcessConfig, error) {
	cfg := &ProcessConfig{
		Handler:    h,
		EventTypes: map[reflect.Type]struct{}{},
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
	t := reflect.TypeOf(m)

	if _, ok := c.cfg.EventTypes[t]; ok {
		panicf(
			`%T.Configure() has already called ProcessConfigurer.RouteEventType(%T)`,
			c.cfg.Handler,
			m,
		)
	}

	c.cfg.EventTypes[t] = struct{}{}
}
