package config

import (
	"context"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/internal/enginekit/message"
)

// ApplicationConfig represents the configuration of an entire Dogma application.
type ApplicationConfig struct {
	// Application is the application that the configuration applies to.
	Application dogma.Application

	// ApplicationName is the application's name, as specified in the dogma.App struct.
	ApplicationName string

	// Handlers is a map of handler name to their respective configuration.
	Handlers map[string]HandlerConfig

	// Routes is map of message type to the names of the handlers that receive
	// messages of that type.
	Routes map[message.Type][]string

	// CommandRoutes is map of command message type to the names of the handler
	// that receives that command type.
	CommandRoutes map[message.Type]string

	// EventRoutes is map of event message type to the names of the handlers that
	// receive events of that type.
	EventRoutes map[message.Type][]string
}

// NewApplicationConfig returns a new application config for the given application.
func NewApplicationConfig(app dogma.Application) (*ApplicationConfig, error) {
	cfg := &ApplicationConfig{
		Application:   app,
		Handlers:      map[string]HandlerConfig{},
		Routes:        map[message.Type][]string{},
		CommandRoutes: map[message.Type]string{},
		EventRoutes:   map[message.Type][]string{},
	}

	c := &applicationConfigurer{
		cfg: cfg,
	}

	if err := catch(func() {
		app.Configure(c)
	}); err != nil {
		return nil, err
	}

	if c.cfg.ApplicationName == "" {
		return nil, errorf(
			"%T.Configure() did not call ApplicationConfigurer.Name()",
			app,
		)
	}
	return cfg, nil
}

// Name returns the application name.
func (c *ApplicationConfig) Name() string {
	return c.ApplicationName
}

// Accept calls v.VisitApplicationConfig(ctx, c).
func (c *ApplicationConfig) Accept(ctx context.Context, v Visitor) error {
	return v.VisitApplicationConfig(ctx, c)
}

// register adds the given handler configuration to the app configuration.
func (c *ApplicationConfig) register(cfg HandlerConfig) {
	n := cfg.Name()

	if x, ok := c.Handlers[n]; ok {
		panicf(
			"%s can not use the handler name %#v, because it is already used by %s",
			cfg.HandlerReflectType(),
			n,
			x.HandlerReflectType(),
		)
	}

	for t := range cfg.CommandTypes() {
		if x, ok := c.CommandRoutes[t]; ok {
			panicf(
				"can not route commands of type %s to %#v because they are already routed to %#v",
				t,
				cfg.Name(),
				x,
			)
		}

		if x, ok := c.EventRoutes[t]; ok {
			if len(x) == 1 {
				panicf(
					"can not route messages of type %s to %#v as commands because they are already routed to %#v as events",
					t,
					cfg.Name(),
					x[0],
				)
			}

			panicf(
				"can not route messages of type %s to %#v as commands because they are already routed to %#v and %d other handler(s) as events",
				t,
				cfg.Name(),
				x[0],
				len(x)-1,
			)
		}
	}

	for t := range cfg.EventTypes() {
		if x, ok := c.CommandRoutes[t]; ok {
			panicf(
				"can not route messages of type %s to %#v as events because they are already routed to %#v as commands",
				t,
				cfg.Name(),
				x,
			)
		}
	}

	c.Handlers[n] = cfg

	for t := range cfg.CommandTypes() {
		c.Routes[t] = append(c.Routes[t], cfg.Name())
		c.CommandRoutes[t] = cfg.Name()
	}

	for t := range cfg.EventTypes() {
		c.Routes[t] = append(c.Routes[t], cfg.Name())
		c.EventRoutes[t] = append(c.EventRoutes[t], cfg.Name())
	}
}

// applicationConfigurer is an implementation of dogma.ApplicationConfigurer
// that builds an ApplicationConfig value.
type applicationConfigurer struct {
	cfg *ApplicationConfig
}

func (c *applicationConfigurer) Name(n string) {
	if c.cfg.ApplicationName != "" {
		panicf(
			`%T.Configure() has already called ApplicationConfigurer.Name(%#v)`,
			c.cfg.Application,
			c.cfg.ApplicationName,
		)
	}

	if strings.TrimSpace(n) == "" {
		panicf(
			`%T.Configure() called ApplicationConfigurer.Name(%#v) with an invalid name`,
			c.cfg.Application,
			n,
		)
	}

	c.cfg.ApplicationName = n
}

// RegisterAggregate configures the engine to route messages to h.
func (c *applicationConfigurer) RegisterAggregate(h dogma.AggregateMessageHandler) {
	cfg, err := NewAggregateConfig(h)
	if err != nil {
		panic(err)
	}

	c.cfg.register(cfg)
}

// RegisterProcess configures the engine to route messages to h.
func (c *applicationConfigurer) RegisterProcess(h dogma.ProcessMessageHandler) {
	cfg, err := NewProcessConfig(h)
	if err != nil {
		panic(err)
	}

	c.cfg.register(cfg)
}

// RegisterIntegration configures the engine to route messages to h.
func (c *applicationConfigurer) RegisterIntegration(h dogma.IntegrationMessageHandler) {
	cfg, err := NewIntegrationConfig(h)
	if err != nil {
		panic(err)
	}

	c.cfg.register(cfg)
}

// RegisterProjection configures the engine to route messages to h.
func (c *applicationConfigurer) RegisterProjection(h dogma.ProjectionMessageHandler) {
	cfg, err := NewProjectionConfig(h)
	if err != nil {
		panic(err)
	}

	c.cfg.register(cfg)
}
