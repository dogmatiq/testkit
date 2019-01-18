package config

import (
	"fmt"
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
func NewProcessConfig(h dogma.ProcessMessageHandler) *ProcessConfig {
	cfg := &ProcessConfig{
		Handler:    h,
		EventTypes: map[reflect.Type]struct{}{},
	}

	c := &processConfigurer{
		cfg: cfg,
	}

	h.Configure(c)

	if c.cfg.HandlerName == "" {
		panic(fmt.Sprintf(
			"%T.Configure() did not call ProcessConfigurer.Name()",
			h,
		))
	}

	if len(c.cfg.EventTypes) == 0 {
		panic(fmt.Sprintf(
			"%T.Configure() did not call ProcessConfigurer.RouteEventType()",
			h,
		))
	}

	return cfg
}

// Name returns the process name.
func (c *ProcessConfig) Name() string {
	return c.HandlerName
}

// Accept calls v.VisitProcessConfig(c).
func (c *ProcessConfig) Accept(v Visitor) {
	v.VisitProcessConfig(c)
}

// processConfigurer is an implementation of dogma.ProcessConfigurer
// that builds an ProcessConfig value.
type processConfigurer struct {
	cfg *ProcessConfig
}

func (c *processConfigurer) Name(n string) {
	if c.cfg.HandlerName != "" {
		panic(fmt.Sprintf(
			`%T.Configure() has already called ProcessConfigurer.Name(#%v)`,
			c.cfg.Handler,
			c.cfg.HandlerName,
		))
	}

	if strings.TrimSpace(n) == "" {
		panic(fmt.Sprintf(
			`%T.Configure() called ProcessConfigurer.Name(#%v) with an invalid name`,
			c.cfg.Handler,
			n,
		))
	}

	c.cfg.HandlerName = n
}

func (c *processConfigurer) RouteEventType(m dogma.Message) {
	t := reflect.TypeOf(m)

	if _, ok := c.cfg.EventTypes[t]; ok {
		panic(fmt.Sprintf(
			`%T.Configure() has already called ProcessConfigurer.RouteEventType(%T)`,
			c.cfg.Handler,
			m,
		))
	}

	c.cfg.EventTypes[t] = struct{}{}
}
