package engine

// Option applies optional engine-wide settings.
type Option func(*Engine)

// WithResetter returns an engine option that registers a reset hook with the
// engine.
//
// fn is a function to be called whenever the engine is reset.
func WithResetter(fn func()) Option {
	if fn == nil {
		panic("fn must not be nil")
	}

	return func(e *Engine) {
		e.resetters = append(e.resetters, fn)
	}
}
