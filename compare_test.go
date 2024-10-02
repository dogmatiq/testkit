package testkit_test

import (
	"context"
	"reflect"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit"
	. "github.com/dogmatiq/testkit/internal/fixtures"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
)

var _ = g.Describe("func DefaultMessageComparator()", func() {
	g.When("the messages are equal", func() {
		g.DescribeTable(
			"it returns true",
			func(a, b dogma.Message) {
				gm.Expect(
					DefaultMessageComparator(a, b),
				).To(gm.BeTrue())
			},
			g.Entry(
				"plain struct",
				CommandA1,
				CommandA1,
			),
			g.Entry(
				"protocol buffers",
				&ProtoMessage{Value: "<value>"},
				&ProtoMessage{Value: "<value>"},
			),
		)

		g.It("ignores unexported fields when comparing protocol buffers messages", func() {
			a := &ProtoMessage{Value: "<value>"}
			b := &ProtoMessage{Value: "<value>"}

			g.By("initializing the unexported fields within one of the messages")
			_ = a.String()

			gm.Expect(
				reflect.DeepEqual(a, b),
			).To(
				gm.BeFalse(),
				"unexported fields do not differ",
			)

			gm.Expect(
				DefaultMessageComparator(a, b),
			).To(gm.BeTrue())
		})
	})

	g.When("the messages are not equal", func() {
		g.DescribeTable(
			"it returns false",
			func(a, b dogma.Message) {
				gm.Expect(
					DefaultMessageComparator(a, b),
				).To(gm.BeFalse())
			},
			g.Entry(
				"different types",
				CommandA1,
				CommandB1,
			),
			g.Entry(
				"protocol buffers",
				&ProtoMessage{Value: "<value-a>"},
				&ProtoMessage{Value: "<value-b>"},
			),
		)
	})
})

var _ = g.Describe("func WithMessageComparator()", func() {
	g.It("configures how messages are compared", func() {
		handler := &IntegrationMessageHandlerStub{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<handler-name>", "7cb41db6-0116-4d03-80d7-277cc391b47e")
				c.Routes(
					dogma.HandlesCommand[CommandStub[TypeA]](),
					dogma.RecordsEvent[EventStub[TypeA]](),
				)
			},
			HandleCommandFunc: func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Command,
			) error {
				s.RecordEvent(EventA1)
				return nil
			},
		}

		app := &ApplicationStub{
			ConfigureFunc: func(c dogma.ApplicationConfigurer) {
				c.Identity("<app>", "477a9515-8318-4229-8f9d-57d84f463cb7")
				c.RegisterIntegration(handler)
			},
		}

		Begin(
			&testingmock.T{},
			app,
			WithMessageComparator(
				func(a, b dogma.Message) bool {
					return true
				},
			),
		).
			EnableHandlers("<handler-name>").
			Expect(
				ExecuteCommand(CommandA1),
				ToRecordEvent(EventA2), // this would fail without our custom comparator
			)
	})
})
