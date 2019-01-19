package engine

import (
	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
)

// Option applies optional settings to an engine.
type Option func(*configurer) error

// Comparator is an engine option that specifies the comparator to use.
func Comparator(c compare.Comparator) Option {
	if c == nil {
		panic("comparator must not be nil")
	}

	return func(cfgr *configurer) error {
		cfgr.comparator = c
		return nil
	}
}

// Renderer is an engine option that specifies the renderer to use.
func Renderer(r render.Renderer) Option {
	if r == nil {
		panic("renderer must not be nil")
	}

	return func(cfgr *configurer) error {
		cfgr.renderer = r
		return nil
	}
}

// DispatchOption applies optional settings while dispatching a message.
type DispatchOption func(*dispatchOptions) error

type dispatchOptions struct {
	observers fact.ObserverSet
}

// Observe is a dispatch option that registers the given observer while the
// message is being dispatched.
func Observe(o fact.Observer) DispatchOption {
	if o == nil {
		panic("observer must not be nil")
	}

	return func(do *dispatchOptions) error {
		do.observers = append(do.observers, o)
		return nil
	}
}
