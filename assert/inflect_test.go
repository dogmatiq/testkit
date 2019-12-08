package assert

import (
	"github.com/dogmatiq/configkit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega" // can't use dot-import because gomega.Assertion conflicts with this package's Assertion type
)

var _ = Describe("func inflect", func() {
	entry := func(r message.Role, in, out string) TableEntry {
		return Entry(
			in+" ("+r.String()+")",
			r,
			in,
			out,
		)
	}

	DescribeTable(
		"returns true if the values have the same type and value",
		func(r message.Role, in, out string) {
			gomega.Expect(inflect(r, in)).To(gomega.Equal(out))
		},
		entry(message.CommandRole, "a <message>", "a command"),
		entry(message.EventRole, "a <message>", "an event"),

		entry(message.CommandRole, "the <messages>", "the commands"),
		entry(message.EventRole, "the <messages>", "the events"),

		entry(message.CommandRole, "<produce> a specific <message>", "execute a specific command"),
		entry(message.EventRole, "<produce> a specific <message>", "record a specific event"),

		entry(message.CommandRole, "the <message> was <produced>", "the command was executed"),
		entry(message.EventRole, "the <message> was <produced>", "the event was recorded"),
	)
})
