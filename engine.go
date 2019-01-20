package dogmatest

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine"
	"github.com/dogmatiq/dogmatest/internal/enginekit/config"
)

// Engine is an alias for engine.Engine.
type Engine = engine.Engine

// NewEngine returns a new test engine.
func NewEngine(app dogma.App, options ...engine.Option) *Engine {
	cfg, err := config.NewAppConfig(app)
	if err != nil {
		panic(err)
	}

	e, err := engine.New(cfg, options...)
	if err != nil {
		panic(err)
	}

	return e
}
