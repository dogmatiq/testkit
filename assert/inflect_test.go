package assert

import (
	. "github.com/dogmatiq/dogmatest/compare"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ Comparator = DefaultComparator{}

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
			Expect(inflect(r, in)).To(Equal(out))
		},
		entry(message.CommandRole, "a <message>", "a command"),
		entry(message.EventRole, "a <message>", "an event"),

		entry(message.CommandRole, "the <messages>", "the commands"),
		entry(message.EventRole, "the <messages>", "the events"),

		entry(message.CommandRole, "<produce> a specific <message>", "execute a specific command"),
		entry(message.EventRole, "<produce> a specific <message>", "record a specific event"),

		entry(message.CommandRole, "the <message> was <produced>", "the command was executed"),
		entry(message.EventRole, "the <message> was <produced>", "the event was recorded"),

		entry(message.CommandRole, "an <other-message>", "an event"),
		entry(message.EventRole, "an <other-message>", "a command"),

		entry(message.CommandRole, "the <other-messages>", "the events"),
		entry(message.EventRole, "the <other-messages>", "the commands"),

		entry(message.CommandRole, "<other-produce> a specific <other-message>", "record a specific event"),
		entry(message.EventRole, "<other-produce> a specific <other-message>", "execute a specific command"),

		entry(message.CommandRole, "the <other-message> was <other-produced>", "the event was recorded"),
		entry(message.EventRole, "the <other-message> was <other-produced>", "the command was executed"),
	)
})
