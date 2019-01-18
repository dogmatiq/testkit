package engine

import (
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/render"
)

// Option is a function that applies optional configuration to an engine.
type Option func(c *configurer, before bool) error

// Comparator is an engine option that specifies the comparator to use.
func Comparator(c compare.Comparator) Option {
	return func(cfgr *configurer, before bool) error {
		if before {
			cfgr.comparator = c
		}

		return nil
	}
}

// Renderer is an engine option that specifies the renderer to use.
func Renderer(r render.Renderer) Option {
	return func(cfgr *configurer, before bool) error {
		if before {
			cfgr.renderer = r
		}

		return nil
	}
}
