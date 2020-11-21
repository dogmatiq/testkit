package report

import (
	"io"

	"github.com/dogmatiq/dapper"
	"github.com/dogmatiq/dogma"
)

// Renderer is an interface for rendering various Dogma values.
type Renderer interface {
	// WriteMessage writes a human-readable representation of v to w.
	//
	// It returns the number of bytes written.
	WriteMessage(w io.Writer, v dogma.Message) (int, error)

	// WriteAggregateRoot writes a human-readable representation of v to w.
	//
	// It returns the number of bytes written.
	WriteAggregateRoot(w io.Writer, v dogma.AggregateRoot) (int, error)

	// WriteProcessRoot writes a human-readable representation of v to w.
	//
	// It returns the number of bytes written.
	WriteProcessRoot(w io.Writer, v dogma.ProcessRoot) (int, error)

	// WriteAggregateMessageHandler writes a human-readable representation of v
	// to w.
	//
	// It returns the number of bytes written.
	WriteAggregateMessageHandler(w io.Writer, v dogma.AggregateMessageHandler) (int, error)

	// WriteProcessMessageHandler writes a human-readable representation of v to
	// w.
	//
	// It returns the number of bytes written.
	WriteProcessMessageHandler(w io.Writer, v dogma.ProcessMessageHandler) (int, error)

	// WriteIntegrationMessageHandler writes a human-readable representation of
	// v to w.
	//
	// It returns the number of bytes written.
	WriteIntegrationMessageHandler(w io.Writer, v dogma.IntegrationMessageHandler) (int, error)

	// WriteProjectionMessageHandler writes a human-readable representation of v
	// to w.
	//
	// It returns the number of bytes written.
	WriteProjectionMessageHandler(w io.Writer, v dogma.ProjectionMessageHandler) (int, error)
}

// DefaultRenderer is the default renderer implementation.
//
// It uses a pretty-printer in a best-effort to render meaningful values.
type DefaultRenderer struct{}

// WriteMessage writes a human-readable representation of v to w.
//
// It returns the number of bytes written.
func (r DefaultRenderer) WriteMessage(w io.Writer, v dogma.Message) (int, error) {
	return r.write(w, v)
}

// WriteAggregateRoot writes a human-readable representation of v to w.
//
// It returns the number of bytes written.
func (r DefaultRenderer) WriteAggregateRoot(w io.Writer, v dogma.AggregateRoot) (int, error) {
	return r.write(w, v)
}

// WriteProcessRoot writes a human-readable representation of v to w.
//
// It returns the number of bytes written.
func (r DefaultRenderer) WriteProcessRoot(w io.Writer, v dogma.ProcessRoot) (int, error) {
	return r.write(w, v)
}

// WriteAggregateMessageHandler writes a human-readable representation of v to
// w.
//
// It returns the number of bytes written.
func (r DefaultRenderer) WriteAggregateMessageHandler(
	w io.Writer,
	v dogma.AggregateMessageHandler,
) (int, error) {
	return r.write(w, v)
}

// WriteProcessMessageHandler writes a human-readable representation of v to w.
//
// It returns the number of bytes written.
func (r DefaultRenderer) WriteProcessMessageHandler(
	w io.Writer,
	v dogma.ProcessMessageHandler,
) (int, error) {
	return r.write(w, v)
}

// WriteIntegrationMessageHandler writes a human-readable representation of v to
// w.
//
// It returns the number of bytes written.
func (r DefaultRenderer) WriteIntegrationMessageHandler(
	w io.Writer,
	v dogma.IntegrationMessageHandler,
) (int, error) {
	return r.write(w, v)
}

// WriteProjectionMessageHandler writes a human-readable representation of v to
// w.
//
// It returns the number of bytes written.
func (r DefaultRenderer) WriteProjectionMessageHandler(
	w io.Writer,
	v dogma.ProjectionMessageHandler,
) (int, error) {
	return r.write(w, v)
}

func (r DefaultRenderer) write(w io.Writer, v interface{}) (int, error) {
	return printer.Write(w, v)
}

// printer is the Dapper printer used to render values.
var printer = dapper.Printer{
	Config: dapper.Config{
		Filters: []dapper.Filter{
			dapper.TimeFilter,
			dapper.DurationFilter,
		},
		OmitPackagePaths: true,
	},
}
