package render

import (
	"io"
	"strings"

	"github.com/dogmatiq/iago"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// WriteDiff renders a human-readable diff of two strings.
func WriteDiff(w io.Writer, a, b string) (n int, err error) {
	defer iago.Recover(&err)

	d := diffmatchpatch.New()

	for _, diff := range d.DiffMain(a, b, false) {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			n += iago.MustWriteString(w, "{+")
			n += iago.MustWriteString(w, text)
			n += iago.MustWriteString(w, "+}")
		case diffmatchpatch.DiffDelete:
			n += iago.MustWriteString(w, "[-")
			n += iago.MustWriteString(w, text)
			n += iago.MustWriteString(w, "-]")
		case diffmatchpatch.DiffEqual:
			n += iago.MustWriteString(w, text)
		}
	}

	return
}

// Diff returns a human-readable diff of two strings.
func Diff(a, b string) string {
	var w strings.Builder
	iago.Must(WriteDiff(&w, a, b))
	return w.String()
}
