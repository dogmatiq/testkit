package assert

import (
	"bytes"
	"io"

	"github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/dogmatest/engine/fact"
	"github.com/dogmatiq/dogmatest/render"
	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/indent"
)

// CompositeAssertion is an assertion that is a container for other assertions.
type CompositeAssertion struct {
	// Title is the title to render in reports.
	Title string

	// SubAssertions is the set of assertions in the container.
	SubAssertions []Assertion

	// Predicate is a function that determines whether or not the assertion passes,
	// based on the number of child assertions that passed.
	//
	// It returns true if the assertion passed, and may optionally return a message
	// to be displayed in either case.
	Predicate func(int) (string, bool)
}

// Notify notifies the assertion of the occurrence of a fact.
func (a *CompositeAssertion) Notify(f fact.Fact) {
	for _, sub := range a.SubAssertions {
		sub.Notify(f)
	}
}

// Begin is called before the message-under-test is dispatched.
func (a *CompositeAssertion) Begin(c compare.Comparator) {
	for _, sub := range a.SubAssertions {
		sub.Begin(c)
	}
}

// End is called after the message-under-test is dispatched.
func (a *CompositeAssertion) End(w io.Writer, r render.Renderer) bool {
	n := 0
	buf := &bytes.Buffer{}

	for _, sub := range a.SubAssertions {
		if sub.End(buf, r) {
			n++
		}
	}

	message, pass := a.Predicate(n)

	writeIcon(w, pass)
	iago.MustWriteString(w, " ")
	iago.MustWriteString(w, a.Title)

	if message != "" {
		iago.MustWriteString(w, " (")
		iago.MustWriteString(w, message)
		iago.MustWriteString(w, " )")
	}

	iago.MustWriteString(w, "\n")
	iago.MustWriteTo(
		indent.NewIndenter(w, nil),
		buf,
	)

	return pass
}
