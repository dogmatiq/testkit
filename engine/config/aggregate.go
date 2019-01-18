package config

import (
	"context"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
)

// AggregateConfig represents the configuration of an aggregate message handler.
type AggregateConfig struct {
	Handler      dogma.AggregateMessageHandler
	HandlerName  string
	CommandTypes map[reflect.Type]struct{}
}

// NewAggregateConfig returns an AggregateConfig for the given handler.
func NewAggregateConfig(h dogma.AggregateMessageHandler) (*AggregateConfig, error) {
	cfg := &AggregateConfig{
		Handler:      h,
		CommandTypes: map[reflect.Type]struct{}{},
	}

	c := &aggregateConfigurer{
		cfg: cfg,
	}

	if err := catch(func() {
		h.Configure(c)
	}); err != nil {
		return nil, err
	}

	if c.cfg.HandlerName == "" {
		return nil, errorf(
			"%T.Configure() did not call AggregateConfigurer.Name()",
			h,
		)
	}

	if len(c.cfg.CommandTypes) == 0 {
		return nil, errorf(
			"%T.Configure() did not call AggregateConfigurer.RouteCommandType()",
			h,
		)
	}

	return cfg, nil
}

// Name returns the aggregate name.
func (c *AggregateConfig) Name() string {
	return c.HandlerName
}

// Accept calls v.VisitAggregateConfig(ctx, c).
func (c *AggregateConfig) Accept(ctx context.Context, v Visitor) error {
	return v.VisitAggregateConfig(ctx, c)
}

// aggregateConfigurer is an implementation of dogma.AggregateConfigurer
// that builds an AggregateConfig value.
type aggregateConfigurer struct {
	cfg *AggregateConfig
}

func (c *aggregateConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panicf(
			`%T.Configure() has already called AggregateConfigurer.Name(%#v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		)
	}

	if strings.TrimSpace(n) == "" {
		panicf(
			`%T.Configure() called AggregateConfigurer.Name(%#v) with an invalid name`,
			c.cfg.Handler,
			n,
		)
	}

	c.cfg.HandlerName = n
}

func (c *aggregateConfigurer) RouteCommandType(m dogma.Message) {
	t := reflect.TypeOf(m)

	if _, ok := c.cfg.CommandTypes[t]; ok {
		panicf(
			`%T.Configure() has already called AggregateConfigurer.RouteCommandType(%T)`,
			c.cfg.Handler,
			m,
		)
	}

	c.cfg.CommandTypes[t] = struct{}{}
}