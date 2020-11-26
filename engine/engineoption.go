package engine

// Option applies optional engine-wide settings.
type Option interface {
	applyEngineOption(*engineOptions)
}

type optionFunc func(*engineOptions)

func (f optionFunc) applyEngineOption(opts *engineOptions) {
	f(opts)
}

// WithResetter returns an engine option that registers a reset hook with the
// engine.
//
// fn is a function to be called whenever the engine is reset.
func WithResetter(fn func()) Option {
	if fn == nil {
		panic("fn must not be nil")
	}

	return optionFunc(func(eo *engineOptions) {
		eo.resetters = append(eo.resetters, fn)
	})
}

// EnableProjectionCompactionDuringHandling returns an engine option that causes
// projection to be compacted in parallel with each event handled.
//
// This option is intended to faciliate testing of compaction logic alongside
// projection building. It is likely not much use when using the engine outside
// of the test runner.
func EnableProjectionCompactionDuringHandling(enabled bool) Option {
	return optionFunc(func(eo *engineOptions) {
		eo.compactDuringHandling = enabled
	})
}

// engineOptions is a container for the options set via Option values.
type engineOptions struct {
	resetters             []func()
	compactDuringHandling bool
}

// newEngineOptions returns a new engineOptions with the given options.
func newEngineOptions(options []Option) *engineOptions {
	eo := &engineOptions{}

	for _, opt := range options {
		opt.applyEngineOption(eo)
	}

	return eo
}
