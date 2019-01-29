package assert

import (
	"io"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/count"
	"github.com/dogmatiq/iago/indent"
)

// writeIcon writes a pass or failure icon to w.
func writeIcon(w io.Writer, pass bool) {
	if pass {
		iago.MustWriteString(w, "✓")
	} else {
		iago.MustWriteString(w, "✗")
	}
}

// byRole returns a message based on a message's role.
func byRole(r message.Role, command, event string) string {
	if r == message.CommandRole {
		return command
	}

	return event
}

// renderDiff returns the diff of a and b.
func renderDiff(a, b string) string {
	var w strings.Builder
	iago.Must(render.WriteDiff(&w, a, b))
	return w.String()
}

// renderMessage returns the representation of m, as per r.
func renderMessage(r render.Renderer, m dogma.Message) string {
	var w strings.Builder
	iago.Must(r.WriteMessage(&w, m))
	return w.String()
}

// report generates an assertion report in the standard format.
type report struct {
	Pass     bool
	Title    string
	SubTitle string
	Note     string
	Hints    []string
	Details  string
}

func (r *report) addHint(h string) {
	r.Hints = append(r.Hints, h)
}

func (r *report) WriteTo(next io.Writer) (_ int64, err error) {
	defer iago.Recover(&err)

	w := count.NewWriter(next)

	writeIcon(w, r.Pass)
	iago.MustWriteByte(w, ' ')
	iago.MustWriteString(w, r.Title)

	if r.SubTitle != "" {
		iago.MustWriteString(w, " (")
		iago.MustWriteString(w, r.SubTitle)
		iago.MustWriteByte(w, ')')
	}

	iago.MustWriteByte(w, '\n')

	indenter := indent.NewIndenter(w, []byte("  | "))
	first := true

	if r.Note != "" {
		if first {
			first = false
			iago.MustWriteByte(w, '\n')
		}

		iago.MustWriteString(indenter, "Note: ")
		iago.MustWriteString(indenter, r.Note)
		iago.MustWriteByte(indenter, '\n')
	}

	if len(r.Hints) > 0 {
		if first {
			first = false
			iago.MustWriteByte(w, '\n')
		} else {
			iago.MustWriteByte(indenter, '\n')
		}

		iago.MustWriteString(indenter, "Suggestions:\n")
		for _, h := range r.Hints {
			iago.MustWriteString(indenter, "  - ")
			iago.MustWriteString(indenter, h)
			iago.MustWriteByte(indenter, '\n')
		}
	}

	if r.Details != "" {
		if first {
			first = false
			iago.MustWriteByte(w, '\n')
		} else {
			iago.MustWriteByte(indenter, '\n')
		}

		iago.MustWriteString(indenter, "Details:\n\n")
		iago.MustWriteString(indenter, r.Details)
	}

	if !first {
		iago.MustWriteByte(w, '\n')
	}

	// TODO(jmalloc): replace with w.Count64()
	// https://github.com/dogmatiq/iago/issues/1
	return int64(w.Count()), nil
}
