package logging_test

import (
	"strings"

	. "github.com/dogmatiq/testkit/fact/internal/logging"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

func describeTable(
	description string,
	fn func(expected string, ids []IconWithLabel, icons []Icon, text []string),
) bool {
	return g.DescribeTable(
		description,
		fn,
		g.Entry(
			"renders a standard log message",
			"= 123  ∵ 456  ⋲ 789  ▼ ↻  <foo> ● <bar>",
			[]IconWithLabel{
				MessageIDIcon.WithLabel("123"),
				CausationIDIcon.WithLabel("456"),
				CorrelationIDIcon.WithLabel("789"),
			},
			[]Icon{
				InboundIcon,
				RetryIcon,
			},
			[]string{
				"<foo>",
				"<bar>",
			},
		),
		g.Entry(
			"renders a hyphen in place of empty labels",
			"= 123  ∵ 456  ⋲ -  ▼    <foo> ● <bar>",
			[]IconWithLabel{
				MessageIDIcon.WithLabel("123"),
				CausationIDIcon.WithLabel("456"),
				CorrelationIDIcon.WithLabel(""),
			},
			[]Icon{
				InboundIcon,
				"",
			},
			[]string{
				"<foo>",
				"<bar>",
			},
		),
		g.Entry(
			"pads empty icons to the same width",
			"= 123  ∵ 456  ⋲ 789  ▼    <foo> ● <bar>",
			[]IconWithLabel{
				MessageIDIcon.WithLabel("123"),
				CausationIDIcon.WithLabel("456"),
				CorrelationIDIcon.WithLabel("789"),
			},
			[]Icon{
				InboundIcon,
				"",
			},
			[]string{
				"<foo>",
				"<bar>",
			},
		),
		g.Entry(
			"skips empty text",
			"= 123  ∵ 456  ⋲ 789  ▼ ↻  <foo> ● <bar>",
			[]IconWithLabel{
				MessageIDIcon.WithLabel("123"),
				CausationIDIcon.WithLabel("456"),
				CorrelationIDIcon.WithLabel("789"),
			},
			[]Icon{
				InboundIcon,
				RetryIcon,
			},
			[]string{
				"<foo>",
				"",
				"<bar>",
			},
		),
	)
}

var _ = describeTable(
	"func String()",
	func(expected string, ids []IconWithLabel, icons []Icon, text []string) {
		gm.Expect(
			String(ids, icons, text...),
		).To(gm.Equal(expected))
	},
)

var _ = describeTable(
	"func Write()",
	func(expected string, ids []IconWithLabel, icons []Icon, text []string) {
		w := &strings.Builder{}

		n, err := Write(w, ids, icons, text...)
		gm.Expect(err).ShouldNot(gm.HaveOccurred())
		gm.Expect(n).To(gm.Equal(len(expected)))

		gm.Expect(w.String()).To(gm.Equal(expected))
	},
)
