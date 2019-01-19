package engine

import (
	"context"
	"reflect"

	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/config"
	"github.com/dogmatiq/dogmatest/engine/controller"
	"github.com/dogmatiq/dogmatest/engine/envelope"
	"github.com/dogmatiq/dogmatest/render"
)

// Engine is an in-memory Dogma engine that is used to execute tests.
type Engine struct {
	roles  map[reflect.Type]envelope.MessageRole
	routes map[reflect.Type][]controller.Controller
}

// New returns a new engine that uses the given app configuration.
func New(
	cfg *config.AppConfig,
	options ...Option,
) (*Engine, error) {
	e := &Engine{
		roles:  map[reflect.Type]envelope.MessageRole{},
		routes: map[reflect.Type][]controller.Controller{},
	}

	cfgr := &configurer{
		engine:     e,
		comparator: compare.DefaultComparator{},
		renderer:   render.DefaultRenderer{},
	}

	ctx := context.Background()

	for _, opt := range options {
		if opt.beforeApp != nil {
			if err := opt.beforeApp(cfgr); err != nil {
				return nil, err
			}
		}
	}

	if err := cfg.Accept(ctx, cfgr); err != nil {
		return nil, err
	}

	for _, opt := range options {
		if opt.afterApp != nil {
			if err := opt.afterApp(cfgr); err != nil {
				return nil, err
			}
		}
	}

	return e, nil
}

// Reset clears the state of the engine.
func (e *Engine) Reset() {
	for _, controllers := range e.routes {
		for _, c := range controllers {
			c.Reset()
		}
	}
}
