package assert

import (
	"io"
	"strings"

	"github.com/dogmatiq/dogma"
	"github.com/dogmatiq/enginekit/message"
	"github.com/dogmatiq/iago/count"
	"github.com/dogmatiq/iago/indent"
	"github.com/dogmatiq/iago/must"
	"github.com/dogmatiq/testkit/render"
)

// writeIcon writes a pass or failure icon to w.
func writeIcon(w io.Writer, pass bool) {
	if pass {
		must.WriteString(w, "✓")
	} else {
		must.WriteString(w, "✗")
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
	must.Must(render.WriteDiff(&w, a, b))
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
	must.Must(r.WriteMessage(&w, m))
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
	defer must.Recover(&err)

	w := count.NewWriter(next)

	writeIcon(w, r.Pass)
	must.WriteByte(w, ' ')
	must.WriteString(w, r.Title)

	if r.SubTitle != "" {
		must.WriteString(w, " (")
		must.WriteString(w, r.SubTitle)
		must.WriteByte(w, ')')
	}

	must.WriteByte(w, '\n')

	indenter := indent.NewIndenter(w, []byte("  | "))
	first := true

	if r.Outcome != "" {
		if first {
			first = false
			must.WriteByte(w, '\n')
		}

		must.WriteString(indenter, "Outcome:\n  ")
		must.WriteString(indenter, r.Outcome)
		must.WriteByte(indenter, '\n')
	}

	if len(r.Suggestions) > 0 {
		if first {
			first = false
			must.WriteByte(w, '\n')
		} else {
			must.WriteByte(indenter, '\n')
		}

		must.WriteString(indenter, "Suggestions:\n")
		for _, s := range r.Suggestions {
			must.WriteString(indenter, "  - ")
			must.WriteString(indenter, s)
			must.WriteByte(indenter, '\n')
		}
	}

	if r.Details != "" {
		if first {
			first = false
			must.WriteByte(w, '\n')
		} else {
			must.WriteByte(indenter, '\n')
		}

		must.WriteString(indenter, "Details:\n")
		must.WriteString(indenter, indent.String(r.Details, "  "))
		must.WriteByte(indenter, '\n')
	}

	if !first {
		must.WriteByte(w, '\n')
	}

	return w.Count64(), nil
}
