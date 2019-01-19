package engine

import (
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/render"
)

// Option applies optional settings to an engine.
type Option struct {
	// beforeApp, if non-nil, is called before applying the app configuration to
	// the engine.
	beforeApp func(*configurer) error

	// afterApp, if non-nil, is called after applying the app configuration to the
	// engine.
	afterApp func(*configurer) error
}

// Comparator is an engine option that specifies the comparator to use.
func Comparator(c compare.Comparator) Option {
	return Option{
		beforeApp: func(cfgr *configurer) error {
			cfgr.comparator = c
			return nil
		},
	}
}

// Renderer is an engine option that specifies the renderer to use.
func Renderer(r render.Renderer) Option {
	return Option{
		beforeApp: func(cfgr *configurer) error {
			cfgr.renderer = r
			return nil
		},
	}
}
