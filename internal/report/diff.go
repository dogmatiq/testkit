package report

import (
	"io"

	"github.com/dogmatiq/iago/must"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// WriteDiff renders a human-readable diff of two strings.
func WriteDiff(w io.Writer, a, b string) {
	d := diffmatchpatch.New()

	for _, diff := range d.DiffMain(a, b, false) {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			must.WriteString(w, "{+")
			must.WriteString(w, text)
			must.WriteString(w, "+}")
		case diffmatchpatch.DiffDelete:
			must.WriteString(w, "[-")
			must.WriteString(w, text)
			must.WriteString(w, "-]")
		case diffmatchpatch.DiffEqual:
			must.WriteString(w, text)
		}
	}
}
