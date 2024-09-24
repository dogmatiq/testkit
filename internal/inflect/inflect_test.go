package inflect_test

import (
	"strings"

	"github.com/dogmatiq/configkit/message"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/internal/inflect"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func Sprint()", func() {
	entry := func(r message.Role, in, out string) g.TableEntry {
		return g.Entry(
			in+" ("+r.String()+")",
			r,
			in,
			out,
		)
	}

	g.DescribeTable(
		"returns a properly inflected string",
		func(r message.Role, in, out string) {
			Expect(Sprint(r, in)).To(Equal(out))

			in = strings.ToUpper(in)
			out = strings.ToUpper(out)
			Expect(Sprint(r, in)).To(Equal(out))
		},
		entry(message.CommandRole, "a <message>", "a command"),
		entry(message.EventRole, "a <message>", "an event"),
		entry(message.TimeoutRole, "a <message>", "a timeout"),

		entry(message.CommandRole, "an <message>", "a command"),
		entry(message.EventRole, "an <message>", "an event"),
		entry(message.TimeoutRole, "an <message>", "a timeout"),

		entry(message.CommandRole, "the <messages>", "the commands"),
		entry(message.EventRole, "the <messages>", "the events"),
		entry(message.TimeoutRole, "the <messages>", "the timeouts"),

		entry(message.CommandRole, "1 <messages>", "1 command"),
		entry(message.EventRole, "1 <messages>", "1 event"),
		entry(message.TimeoutRole, "1 <messages>", "1 timeout"),

		entry(message.CommandRole, "21 <messages>", "21 commands"),
		entry(message.EventRole, "21 <messages>", "21 events"),
		entry(message.TimeoutRole, "21 <messages>", "21 timeouts"),

		entry(message.CommandRole, "only 1 <messages>", "only 1 command"),
		entry(message.EventRole, "only 1 <messages>", "only 1 event"),
		entry(message.TimeoutRole, "only 1 <messages>", "only 1 timeout"),

		entry(message.CommandRole, "only 21 <messages>", "only 21 commands"),
		entry(message.EventRole, "only 21 <messages>", "only 21 events"),
		entry(message.TimeoutRole, "only 21 <messages>", "only 21 timeouts"),

		entry(message.CommandRole, "<produce> a specific <message>", "execute a specific command"),
		entry(message.EventRole, "<produce> a specific <message>", "record a specific event"),
		entry(message.TimeoutRole, "<produce> a specific <message>", "schedule a specific timeout"),

		entry(message.CommandRole, "the <message> was <produced>", "the command was executed"),
		entry(message.EventRole, "the <message> was <produced>", "the event was recorded"),
		entry(message.TimeoutRole, "the <message> was <produced>", "the timeout was scheduled"),

		entry(message.CommandRole, "via a <dispatcher>", "via a dogma.CommandExecutor"),
		entry(message.EventRole, "via a <dispatcher>", "via a dogma.EventRecorder"),
	)
})

var _ = g.Describe("func Sprintf()", func() {
	g.It("returns the inflected and substituted string", func() {
		Expect(
			Sprintf(
				message.CommandRole,
				"the %T <message>",
				CommandA1,
			),
		).To(Equal("the stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA] command"))
	})
})

var _ = g.Describe("func Error()", func() {
	g.It("returns an error with the inflected message", func() {
		Expect(
			Error(
				message.CommandRole,
				"the <message>",
			),
		).To(MatchError("the command"))
	})
})

var _ = g.Describe("func Errorf()", func() {
	g.It("returns an error with the inflected and substituted message", func() {
		Expect(
			Errorf(
				message.CommandRole,
				"the %T <message>",
				CommandA1,
			),
		).To(MatchError("the stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA] command"))
	})
})
