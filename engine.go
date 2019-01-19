package dogmatest

import (
	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/engine"
	"github.com/dogmatiq/dogmatest/engine/config"
)

// NewEngine returns a new test engine.
func NewEngine(app dogma.App, options ...engine.Option) *engine.Engine {
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
