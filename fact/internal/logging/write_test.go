package logging_test

import (
	"strings"
	"testing"

	. "github.com/dogmatiq/testkit/fact/internal/logging"
	"github.com/dogmatiq/testkit/internal/test"
)

func TestString(t *testing.T) {
	for _, c := range writeTestCases {
		t.Run(c.Name, func(t *testing.T) {
			test.Expect(
				t,
				"unexpected string",
				String(c.Ids, c.Icons, c.Text...),
				c.Expected,
			)
		})
	}
}

func TestWrite(t *testing.T) {
	for _, c := range writeTestCases {
		t.Run(c.Name, func(t *testing.T) {
			w := &strings.Builder{}

			n, err := Write(w, c.Ids, c.Icons, c.Text...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			test.Expect(t, "unexpected byte count", n, len(c.Expected))
			test.Expect(t, "unexpected output", w.String(), c.Expected)
		})
	}
}

var writeTestCases = []struct {
	Name     string
	Expected string
	Ids      []IconWithLabel
	Icons    []Icon
	Text     []string
}{
	{
		Name:     "renders a standard log message",
		Expected: "= 123  ∵ 456  ⋲ 789  ▼ ↻  <foo> ● <bar>",
		Ids: []IconWithLabel{
			MessageIDIcon.WithLabel("123"),
			CausationIDIcon.WithLabel("456"),
			CorrelationIDIcon.WithLabel("789"),
		},
		Icons: []Icon{InboundIcon, RetryIcon},
		Text:  []string{"<foo>", "<bar>"},
	},
	{
		Name:     "renders a hyphen in place of empty labels",
		Expected: "= 123  ∵ 456  ⋲ -  ▼    <foo> ● <bar>",
		Ids: []IconWithLabel{
			MessageIDIcon.WithLabel("123"),
			CausationIDIcon.WithLabel("456"),
			CorrelationIDIcon.WithLabel(""),
		},
		Icons: []Icon{InboundIcon, ""},
		Text:  []string{"<foo>", "<bar>"},
	},
	{
		Name:     "pads empty icons to the same width",
		Expected: "= 123  ∵ 456  ⋲ 789  ▼    <foo> ● <bar>",
		Ids: []IconWithLabel{
			MessageIDIcon.WithLabel("123"),
			CausationIDIcon.WithLabel("456"),
			CorrelationIDIcon.WithLabel("789"),
		},
		Icons: []Icon{InboundIcon, ""},
		Text:  []string{"<foo>", "<bar>"},
	},
	{
		Name:     "skips empty text",
		Expected: "= 123  ∵ 456  ⋲ 789  ▼ ↻  <foo> ● <bar>",
		Ids: []IconWithLabel{
			MessageIDIcon.WithLabel("123"),
			CausationIDIcon.WithLabel("456"),
			CorrelationIDIcon.WithLabel("789"),
		},
		Icons: []Icon{InboundIcon, RetryIcon},
		Text:  []string{"<foo>", "", "<bar>"},
	},
}
