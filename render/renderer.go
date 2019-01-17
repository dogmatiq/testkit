package render

import (
	"io"

	"github.com/dogmatiq/dapper"
	"github.com/dogmatiq/dogma"
)

// Renderer is an interface for rendering various Dogma values.
type Renderer interface {
	// WriteMessage writes a human-readable representation of v to w.
	// It returns the number of bytes written.
	WriteMessage(w io.Writer, v dogma.Message) (int, error)

	// WriteAggregateRoot writes a human-readable representation of v to w.
	// It returns the number of bytes written.
	WriteAggregateRoot(w io.Writer, v dogma.AggregateRoot) (int, error)

	// WriteProcessRoot writes a human-readable representation of v to w.
	// It returns the number of bytes written.
	WriteProcessRoot(w io.Writer, v dogma.ProcessRoot) (int, error)
}

// DefaultRenderer is the default renderer implementation.
//
// It uses a pretty-printer in a best-effort to render meaningful values.
type DefaultRenderer struct{}

// WriteMessage writes a human-readable representation of v to w.
// It returns the number of bytes written.
func (r DefaultRenderer) WriteMessage(w io.Writer, v dogma.Message) (int, error) {
	return r.write(w, v)
}

// WriteAggregateRoot writes a human-readable representation of v to w.
// It returns the number of bytes written.
func (r DefaultRenderer) WriteAggregateRoot(w io.Writer, v dogma.AggregateRoot) (int, error) {
	return r.write(w, v)
}

// WriteProcessRoot writes a human-readable representation of v to w.
// It returns the number of bytes written.
func (r DefaultRenderer) WriteProcessRoot(w io.Writer, v dogma.ProcessRoot) (int, error) {
	return r.write(w, v)
}

func (r DefaultRenderer) write(w io.Writer, v interface{}) (int, error) {
	return dapper.Write(w, v)
}
