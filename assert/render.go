package assert

import (
	"io"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/count"
	"github.com/dogmatiq/iago/indent"
	"github.com/dogmatiq/testkit/render"
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

// renderMessageDiff returns the diff of a and b.
func renderMessageDiff(r render.Renderer, a, b dogma.Message) string {
	return renderDiff(
		renderMessage(r, a), renderMessage(r, b),
	)
}

// renderMessage returns the representation of m, as per r.
func renderMessage(r render.Renderer, m dogma.Message) string {
	var w strings.Builder
	iago.Must(r.WriteMessage(&w, m))
	return w.String()
}

// report generates an assertion report in the standard format.
type report struct {
	Pass        bool
	Title       string
	SubTitle    string
	Outcome     string
	Suggestions []string
	Details     string
}

func (r *report) suggest(h string) {
	r.Suggestions = append(r.Suggestions, h)
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

	if r.Outcome != "" {
		if first {
			first = false
			iago.MustWriteByte(w, '\n')
		}

		iago.MustWriteString(indenter, "Outcome:\n  ")
		iago.MustWriteString(indenter, r.Outcome)
		iago.MustWriteByte(indenter, '\n')
	}

	if len(r.Suggestions) > 0 {
		if first {
			first = false
			iago.MustWriteByte(w, '\n')
		} else {
			iago.MustWriteByte(indenter, '\n')
		}

		iago.MustWriteString(indenter, "Suggestions:\n")
		for _, s := range r.Suggestions {
			iago.MustWriteString(indenter, "  - ")
			iago.MustWriteString(indenter, s)
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

		iago.MustWriteString(indenter, "Details:\n")
		iago.MustWriteString(indenter, indent.String(r.Details, "  "))
		iago.MustWriteByte(indenter, '\n')
	}

	if !first {
		iago.MustWriteByte(w, '\n')
	}

	// TODO(jmalloc): replace with w.Count64()
	// https://github.com/dogmatiq/iago/issues/1
	return int64(w.Count()), nil
}
