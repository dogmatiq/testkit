package config

import (
	"context"
	"reflect"
	"strings"

	"github.com/dogmatiq/dogma"
)

// AppConfig represents the configuration of an entire Dogma application.Config.
type AppConfig struct {
	AppName       string
	Handlers      map[string]Config
	Routes        map[reflect.Type][]string
	CommandRoutes map[reflect.Type]string
	EventRoutes   map[reflect.Type][]string
}

// NewAppConfig returns a new application config for the given application.
func NewAppConfig(app dogma.App) (*AppConfig, error) {
	if strings.TrimSpace(app.Name) == "" {
		panic("application name must not be empty")
	}

	cfg := &AppConfig{
		AppName:       app.Name,
		Handlers:      map[string]Config{},
		Routes:        map[reflect.Type][]string{},
		CommandRoutes: map[reflect.Type]string{},
		EventRoutes:   map[reflect.Type][]string{},
	}

	ctx := context.Background()
	v := &registerer{cfg}

	for _, h := range app.Aggregates {
		c, err := NewAggregateConfig(h)
		if err != nil {
			return nil, err
		}

		if err := c.Accept(ctx, v); err != nil {
			return nil, err
		}
	}

	for _, h := range app.Processes {
		c, err := NewProcessConfig(h)
		if err != nil {
			return nil, err
		}

		if err := c.Accept(ctx, v); err != nil {
			return nil, err
		}
	}

	for _, h := range app.Integrations {
		c, err := NewIntegrationConfig(h)
		if err != nil {
			return nil, err
		}

		if err := c.Accept(ctx, v); err != nil {
			return nil, err
		}
	}

	for _, h := range app.Projections {
		c, err := NewProjectionConfig(h)
		if err != nil {
			return nil, err
		}

		if err := c.Accept(ctx, v); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// Name returns the application name.
func (c *AppConfig) Name() string {
	return c.AppName
}

// Accept calls v.VisitAppConfig(ctx, c).
func (c *AppConfig) Accept(ctx context.Context, v Visitor) error {
	return v.VisitAppConfig(ctx, c)
}

func (c *AppConfig) registerHandlerConfig(
	cfg Config,
	commandTypes map[reflect.Type]struct{},
	eventTypes map[reflect.Type]struct{},
) error {
	n := cfg.Name()

	if x, ok := c.Handlers[n]; ok {
		return errorf(
			"%s can not use the handler name %#v, because it is already used by %s",
			handlerType(cfg),
			n,
			handlerType(x),
		)
	}

	for t := range commandTypes {
		if x, ok := c.CommandRoutes[t]; ok {
			return errorf(
				"can not route commands of type %s to %#v because they are already routed to %#v",
				t,
				cfg.Name(),
				x,
			)
		}

		if x, ok := c.EventRoutes[t]; ok {
			return errorf(
				"can not route messages of type %s to %#v as commands because they are already routed to %#v and %d other handlers as events",
				t,
				cfg.Name(),
				x[0],
				len(x)-1,
			)
		}
	}

	for t := range eventTypes {
		if x, ok := c.CommandRoutes[t]; ok {
			return errorf(
				"can not route messages of type %s to %#v as events because they are already routed to %#v as commands",
				t,
				cfg.Name(),
				x,
			)
		}
	}

	c.Handlers[n] = cfg

	for t := range commandTypes {
		c.Routes[t] = append(c.Routes[t], cfg.Name())
		c.CommandRoutes[t] = cfg.Name()
	}

	for t := range eventTypes {
		c.Routes[t] = append(c.Routes[t], cfg.Name())
		c.EventRoutes[t] = append(c.EventRoutes[t], cfg.Name())
	}

	return nil
}

// registerer is a visitor that registers handler configs with the app config.
type registerer struct {
	cfg *AppConfig
}

func (r *registerer) VisitAppConfig(context.Context, *AppConfig) error {
	panic("not implemented")
}

func (r *registerer) VisitAggregateConfig(_ context.Context, cfg *AggregateConfig) error {
	return r.cfg.registerHandlerConfig(cfg, cfg.CommandTypes, nil)
}

func (r *registerer) VisitProcessConfig(_ context.Context, cfg *ProcessConfig) error {
	return r.cfg.registerHandlerConfig(cfg, nil, cfg.EventTypes)
}

// VisitIntegrationConfig merges cfg with c.
func (r *registerer) VisitIntegrationConfig(_ context.Context, cfg *IntegrationConfig) error {
	return r.cfg.registerHandlerConfig(cfg, cfg.CommandTypes, nil)
}

// VisitProjectionConfig merges cfg with c.
func (r *registerer) VisitProjectionConfig(_ context.Context, cfg *ProjectionConfig) error {
	return r.cfg.registerHandlerConfig(cfg, nil, cfg.EventTypes)
}
