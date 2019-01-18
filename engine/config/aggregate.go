package config

import (
	"fmt"
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
func NewAggregateConfig(h dogma.AggregateMessageHandler) *AggregateConfig {
	cfg := &AggregateConfig{
		Handler:      h,
		CommandTypes: map[reflect.Type]struct{}{},
	}

	c := &aggregateConfigurer{
		cfg: cfg,
	}

	h.Configure(c)

	if c.cfg.HandlerName == "" {
		panic(fmt.Sprintf(
			"%T.Configure() did not call AggregateConfigurer.Name()",
			h,
		))
	}

	if len(c.cfg.CommandTypes) == 0 {
		panic(fmt.Sprintf(
			"%T.Configure() did not call AggregateConfigurer.RouteCommandType()",
			h,
		))
	}

	return cfg
}

// Name returns the aggregate name.
func (c *AggregateConfig) Name() string {
	return c.HandlerName
}

// Accept calls v.VisitAggregateConfig(c).
func (c *AggregateConfig) Accept(v Visitor) {
	v.VisitAggregateConfig(c)
}

// aggregateConfigurer is an implementation of dogma.AggregateConfigurer
// that builds an AggregateConfig value.
type aggregateConfigurer struct {
	cfg *AggregateConfig
}

func (c *aggregateConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panic(fmt.Sprintf(
			`%T.Configure() has already called AggregateConfigurer.Name(#%v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		))
	}

	if strings.TrimSpace(n) == "" {
		panic(fmt.Sprintf(
			`%T.Configure() called AggregateConfigurer.Name(#%v) with an invalid name`,
			c.cfg.Handler,
			n,
		))
	}

	c.cfg.HandlerName = n
}

func (c *aggregateConfigurer) RouteCommandType(m dogma.Message) {
	t := reflect.TypeOf(m)

	if _, ok := c.cfg.CommandTypes[t]; ok {
		panic(fmt.Sprintf(
			`%T.Configure() has already called AggregateConfigurer.RouteCommandType(%T)`,
			c.cfg.Handler,
			m,
		))
	}

	c.cfg.CommandTypes[t] = struct{}{}
}
