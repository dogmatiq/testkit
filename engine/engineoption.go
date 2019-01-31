package engine

// Option applies optional engine-wide settings.
type Option func(*engineOptions)

// WithResetter returns an engine option that registers a reset hook with the
// engine.
//
// fn is a function to be called whenever the engine is reset.
func WithResetter(fn func()) Option {
	if fn == nil {
		panic("fn must not be nil")
	}

	return func(eo *engineOptions) {
		eo.resetters = append(eo.resetters, fn)
	}
}

// engineOptions is a container for the options set via Option values.
type engineOptions struct {
	resetters []func()
}

// newEngineOptions returns a new engineOptions with the given options.
func newEngineOptions(options []Option) *engineOptions {
	eo := &engineOptions{}

	for _, opt := range options {
		opt(eo)
	}

	return eo
}
