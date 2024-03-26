package testkit_test

import (
	"context"
	"reflect"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/fixtures"
	"github.com/dogmatiq/testkit/internal/testingmock"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("func DefaultMessageComparator()", func() {
	g.When("the messages are equal", func() {
		g.DescribeTable(
			"it returns true",
			func(a, b dogma.Message) {
				Expect(
					DefaultMessageComparator(a, b),
				).To(BeTrue())
			},
			g.Entry(
				"plain struct",
				MessageA1,
				MessageA1,
			),
			g.Entry(
				"protocol buffers",
				&fixtures.ProtoMessage{Value: "<value>"},
				&fixtures.ProtoMessage{Value: "<value>"},
			),
		)

		g.It("ignores unexported fields when comparing protocol buffers messages", func() {
			a := &fixtures.ProtoMessage{Value: "<value>"}
			b := &fixtures.ProtoMessage{Value: "<value>"}

			g.By("initializing the unexported fields within one of the messages")
			_ = a.String()

			Expect(
				reflect.DeepEqual(a, b),
			).To(
				BeFalse(),
				"unexported fields do not differ",
			)

			Expect(
				DefaultMessageComparator(a, b),
			).To(BeTrue())
		})
	})

	g.When("the messages are not equal", func() {
		g.DescribeTable(
			"it returns false",
			func(a, b dogma.Message) {
				Expect(
					DefaultMessageComparator(a, b),
				).To(BeFalse())
			},
			g.Entry(
				"different types",
				MessageA1,
				MessageB1,
			),
			g.Entry(
				"protocol buffers",
				&fixtures.ProtoMessage{Value: "<value-a>"},
				&fixtures.ProtoMessage{Value: "<value-b>"},
			),
		)
	})
})

var _ = g.Describe("func WithMessageComparator()", func() {
	g.It("configures how messages are compared", func() {
		handler := &IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<handler-name>", "7cb41db6-0116-4d03-80d7-277cc391b47e")
				c.Routes(
					dogma.HandlesCommand[MessageC](),
					dogma.RecordsEvent[MessageE](),
				)
			},
			HandleCommandFunc: func(
				_ context.Context,
				s dogma.IntegrationCommandScope,
				_ dogma.Message,
			) error {
				s.RecordEvent(MessageE1)
				return nil
			},
		}

		app := &Application{
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
				ExecuteCommand(MessageC1),
				ToRecordEvent(MessageE2), // this would fail without our custom comparator
			)
	})
})
