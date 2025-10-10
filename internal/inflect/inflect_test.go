package inflect_test

import (
	"strings"

	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	"github.com/dogmatiq/enginekit/message"
	. "github.com/dogmatiq/testkit/internal/inflect"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func Sprint()", func() {
	entry := func(k message.Kind, in, out string) g.TableEntry {
		return g.Entry(
			in+" ("+k.String()+")",
			k,
			in,
			out,
		)
	}

	g.DescribeTable(
		"returns a properly inflected string",
		func(k message.Kind, in, out string) {
			gm.Expect(Sprint(k, in)).To(gm.Equal(out))

			in = strings.ToUpper(in)
			out = strings.ToUpper(out)
			gm.Expect(Sprint(k, in)).To(gm.Equal(out))
		},
		entry(message.CommandKind, "a <message>", "a command"),
		entry(message.EventKind, "a <message>", "an event"),
		entry(message.TimeoutKind, "a <message>", "a timeout"),

		entry(message.CommandKind, "an <message>", "a command"),
		entry(message.EventKind, "an <message>", "an event"),
		entry(message.TimeoutKind, "an <message>", "a timeout"),

		entry(message.CommandKind, "the <messages>", "the commands"),
		entry(message.EventKind, "the <messages>", "the events"),
		entry(message.TimeoutKind, "the <messages>", "the timeouts"),

		entry(message.CommandKind, "1 <messages>", "1 command"),
		entry(message.EventKind, "1 <messages>", "1 event"),
		entry(message.TimeoutKind, "1 <messages>", "1 timeout"),

		entry(message.CommandKind, "21 <messages>", "21 commands"),
		entry(message.EventKind, "21 <messages>", "21 events"),
		entry(message.TimeoutKind, "21 <messages>", "21 timeouts"),

		entry(message.CommandKind, "only 1 <messages>", "only 1 command"),
		entry(message.EventKind, "only 1 <messages>", "only 1 event"),
		entry(message.TimeoutKind, "only 1 <messages>", "only 1 timeout"),

		entry(message.CommandKind, "only 21 <messages>", "only 21 commands"),
		entry(message.EventKind, "only 21 <messages>", "only 21 events"),
		entry(message.TimeoutKind, "only 21 <messages>", "only 21 timeouts"),

		entry(message.CommandKind, "<produce> a specific <message>", "execute a specific command"),
		entry(message.EventKind, "<produce> a specific <message>", "record a specific event"),
		entry(message.TimeoutKind, "<produce> a specific <message>", "schedule a specific timeout"),

		entry(message.CommandKind, "the <message> was <produced>", "the command was executed"),
		entry(message.EventKind, "the <message> was <produced>", "the event was recorded"),
		entry(message.TimeoutKind, "the <message> was <produced>", "the timeout was scheduled"),

		entry(message.CommandKind, "via a <dispatcher>", "via a dogma.CommandExecutor"),
		entry(message.EventKind, "via a <dispatcher>", "via a dogma.EventRecorder"),
	)
})

var _ = g.Describe("func Sprintf()", func() {
	g.It("returns the inflected and substituted string", func() {
		gm.Expect(
			Sprintf(
				message.CommandKind,
				"the %T <message>",
				CommandA1,
			),
		).To(gm.Equal("the *stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA] command"))
	})
})

var _ = g.Describe("func Error()", func() {
	g.It("returns an error with the inflected message", func() {
		gm.Expect(
			Error(
				message.CommandKind,
				"the <message>",
			),
		).To(gm.MatchError("the command"))
	})
})

var _ = g.Describe("func Errorf()", func() {
	g.It("returns an error with the inflected and substituted message", func() {
		gm.Expect(
			Errorf(
				message.CommandKind,
				"the %T <message>",
				CommandA1,
			),
		).To(gm.MatchError("the *stubs.CommandStub[github.com/dogmatiq/enginekit/enginetest/stubs.TypeA] command"))
	})
})
