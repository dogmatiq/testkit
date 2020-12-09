package report

import (
	"io"

	"github.com/dogmatiq/iago/must"
)

// RenderText renders a report in plain-text format.
func RenderText(w io.Writer, r Report) (err error) {
	defer must.Recover(&err)

	must.WriteString(w, "--- ")
	must.WriteString(w, r.Caption.String())
	must.WriteString(w, " ---\n")

	return nil
}
