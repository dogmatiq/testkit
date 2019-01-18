package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
)

// IntegrationConfig represents the configuration of an integration message handler.
type IntegrationConfig struct {
	Handler      dogma.IntegrationMessageHandler
	HandlerName  string
	CommandTypes map[reflect.Type]struct{}
}

// NewIntegrationConfig returns an IntegrationConfig for the given handler.
func NewIntegrationConfig(h dogma.IntegrationMessageHandler) *IntegrationConfig {
	cfg := &IntegrationConfig{
		Handler:      h,
		CommandTypes: map[reflect.Type]struct{}{},
	}

	c := &integrationConfigurer{
		cfg: cfg,
	}

	h.Configure(c)

	if c.cfg.HandlerName == "" {
		panic(fmt.Sprintf(
			"%T.Configure() did not call IntegrationConfigurer.Name()",
			h,
		))
	}

	if len(c.cfg.CommandTypes) == 0 {
		panic(fmt.Sprintf(
			"%T.Configure() did not call IntegrationConfigurer.RouteCommandType()",
			h,
		))
	}

	return cfg
}

// Name returns the integration name.
func (c *IntegrationConfig) Name() string {
	return c.HandlerName
}

// Accept calls v.VisitIntegrationConfig(c).
func (c *IntegrationConfig) Accept(v Visitor) {
	v.VisitIntegrationConfig(c)
}

// integrationConfigurer is an implementation of dogma.integrationConfigurer
// that builds an IntegrationConfig value.
type integrationConfigurer struct {
	cfg *IntegrationConfig
}

func (c *integrationConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panic(fmt.Sprintf(
			`%T.Configure() has already called IntegrationConfigurer.Name(#%v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		))
	}

	if strings.TrimSpace(n) == "" {
		panic(fmt.Sprintf(
			`%T.Configure() called IntegrationConfigurer.Name(#%v) with an invalid name`,
			c.cfg.Handler,
			n,
		))
	}

	c.cfg.HandlerName = n
}

func (c *integrationConfigurer) RouteCommandType(m dogma.Message) {
	t := reflect.TypeOf(m)

	if _, ok := c.cfg.CommandTypes[t]; ok {
		panic(fmt.Sprintf(
			`%T.Configure() has already called IntegrationConfigurer.RouteCommandType(%T)`,
			c.cfg.Handler,
			m,
		))
	}

	c.cfg.CommandTypes[t] = struct{}{}
}
