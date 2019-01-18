package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
)

// AppConfig represents the configuration of an entire Dogma application.Config.
type AppConfig struct {
	AppName       string
	Handlers      map[string]Config
	Routes        map[reflect.Type][]Config
	CommandRoutes map[reflect.Type]Config
	EventRoutes   map[reflect.Type][]Config
}

// NewAppConfig returns a new application config for the given application.
// It panics if the app's configuration can not be validated.
func NewAppConfig(app dogma.App) *AppConfig {
	if strings.TrimSpace(app.Name) == "" {
		panic("application name must not be empty")
	}

	cfg := &AppConfig{
		AppName:       app.Name,
		Handlers:      map[string]Config{},
		Routes:        map[reflect.Type][]Config{},
		CommandRoutes: map[reflect.Type]Config{},
		EventRoutes:   map[reflect.Type][]Config{},
	}

	v := &registerer{cfg}

	for _, h := range app.Aggregates {
		NewAggregateConfig(h).Accept(v)
	}

	for _, h := range app.Processes {
		NewProcessConfig(h).Accept(v)
	}

	for _, h := range app.Integrations {
		NewIntegrationConfig(h).Accept(v)
	}

	for _, h := range app.Projections {
		NewProjectionConfig(h).Accept(v)
	}

	return cfg
}

// Name returns the application name.
func (c *AppConfig) Name() string {
	return c.AppName
}

// Accept calls v.VisitAppConfig(c).
func (c *AppConfig) Accept(v Visitor) {
	v.VisitAppConfig(c)
}

func (c *AppConfig) registerHandlerConfig(
	cfg Config,
	commandTypes map[reflect.Type]struct{},
	eventTypes map[reflect.Type]struct{},
) {
	n := cfg.Name()

	if x, ok := c.Handlers[n]; ok {
		panic(fmt.Sprintf(
			"%s can not use the handler name %#v, because it is already used by %s",
			handlerType(cfg),
			n,
			handlerType(x),
		))
	}

	for t := range commandTypes {
		if x, ok := c.CommandRoutes[t]; ok {
			panic(fmt.Sprintf(
				"can not route commands of type %s to %#v because they are already routed to %#v",
				t,
				cfg.Name(),
				x.Name(),
			))
		}

		if x, ok := c.EventRoutes[t]; ok {
			panic(fmt.Sprintf(
				"can not route messages of type %s to %#v as commands because they are already routed to %#v and %d other handlers as events",
				t,
				cfg.Name(),
				x[0].Name(),
				len(x)-1,
			))
		}
	}

	for t := range eventTypes {
		if x, ok := c.CommandRoutes[t]; ok {
			panic(fmt.Sprintf(
				"can not route messages of type %s to %#v as events because they are already routed to %#v as commands",
				t,
				cfg.Name(),
				x.Name(),
			))
		}
	}

	c.Handlers[n] = cfg

	for t := range commandTypes {
		c.Routes[t] = append(c.Routes[t], cfg)
		c.CommandRoutes[t] = cfg
	}

	for t := range eventTypes {
		c.Routes[t] = append(c.Routes[t], cfg)
		c.EventRoutes[t] = append(c.EventRoutes[t], cfg)
	}
}

// registerer is a visitor that registers handler configs with the app config.
type registerer struct {
	cfg *AppConfig
}

func (r *registerer) VisitAppConfig(*AppConfig) {
	panic("not implemented")
}

func (r *registerer) VisitAggregateConfig(cfg *AggregateConfig) {
	r.cfg.registerHandlerConfig(cfg, cfg.CommandTypes, nil)
}

func (r *registerer) VisitProcessConfig(cfg *ProcessConfig) {
	r.cfg.registerHandlerConfig(cfg, nil, cfg.EventTypes)
}

// VisitIntegrationConfig merges cfg with c.
func (r *registerer) VisitIntegrationConfig(cfg *IntegrationConfig) {
	r.cfg.registerHandlerConfig(cfg, cfg.CommandTypes, nil)
}

// VisitProjectionConfig merges cfg with c.
func (r *registerer) VisitProjectionConfig(cfg *ProjectionConfig) {
	r.cfg.registerHandlerConfig(cfg, nil, cfg.EventTypes)
}
