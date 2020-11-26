package report

import (
	"io"
	"strings"

	"github.com/dogmatiq/iago/must"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// WriteDiff renders a human-readable diff of two strings.
func WriteDiff(w io.Writer, a, b string) (n int, err error) {
	defer must.Recover(&err)

	d := diffmatchpatch.New()

	for _, diff := range d.DiffMain(a, b, false) {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			n += must.WriteString(w, "{+")
			n += must.WriteString(w, text)
			n += must.WriteString(w, "+}")
		case diffmatchpatch.DiffDelete:
			n += must.WriteString(w, "[-")
			n += must.WriteString(w, text)
			n += must.WriteString(w, "-]")
		case diffmatchpatch.DiffEqual:
			n += must.WriteString(w, text)
		}
	}

	return
}

// Diff returns a human-readable diff of two strings.
func Diff(a, b string) string {
	var w strings.Builder
	must.Must(WriteDiff(&w, a, b))
	return w.String()
}
