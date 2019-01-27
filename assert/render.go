package assert

import (
	"io"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/render"

	"github.com/dogmatiq/enginekit/message"

	"github.com/dogmatiq/iago"
)

// writeIcon writes a pass or failure icon to w.
func writeIcon(w io.Writer, pass bool) {
	if pass {
		iago.MustWriteString(w, "✓")
	} else {
		iago.MustWriteString(w, "✗")
	}
}

// writeHintByRole writes one of three given messages depending on whether the a
// message role is the same as some expected role, or if it differs, whether the
// expected role is CommandRole or EventType.
func writeHintByRole(
	w io.Writer,
	expected, actual message.Role,
	same, command, event string,
) {
	m := ""

	switch actual {
	case expected:
		m = same
	case message.CommandRole:
		m = command
	case message.EventRole:
		m = event
	}

	if m == "" {
		panic("internal assert error: no message provided")
	}

	iago.MustWriteString(w, "Hint: ")
	iago.MustWriteString(w, m)
	iago.MustWriteString(w, "\n")
}

// writeDiff writes the diff of two messages to w.
func writeDiff(
	w io.Writer,
	r render.Renderer,
	a, b dogma.Message,
) {
	var aw, bw strings.Builder

	iago.Must(
		r.WriteMessage(&aw, a),
	)

	iago.Must(
		r.WriteMessage(&bw, b),
	)

	iago.Must(
		render.WriteDiff(
			w,
			aw.String(),
			bw.String(),
		),
	)
}
