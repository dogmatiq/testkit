package testkit_test

import (
	"context"
	"reflect"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/fixtures"
	"github.com/dogmatiq/testkit/internal/testingmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("func DefaultMessageComparator()", func() {
	When("the messages are equal", func() {
		DescribeTable(
			"it returns true",
			func(a, b dogma.Message) {
				Expect(
					DefaultMessageComparator(a, b),
				).To(BeTrue())
			},
			Entry(
				"plain struct",
				MessageA1,
				MessageA1,
			),
			Entry(
				"protocol buffers",
				&fixtures.ProtoMessage{Value: "<value>"},
				&fixtures.ProtoMessage{Value: "<value>"},
			),
		)

		It("ignores unexported fields when comparing protocol buffers messages", func() {
			a := &fixtures.ProtoMessage{Value: "<value>"}
			b := &fixtures.ProtoMessage{Value: "<value>"}

			By("initializing the unexported fields within one of the messages")
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

	When("the messages are not equal", func() {
		DescribeTable(
			"it returns false",
			func(a, b dogma.Message) {
				Expect(
					DefaultMessageComparator(a, b),
				).To(BeFalse())
			},
			Entry(
				"different types",
				MessageA1,
				MessageB1,
			),
			Entry(
				"protocol buffers",
				&fixtures.ProtoMessage{Value: "<value-a>"},
				&fixtures.ProtoMessage{Value: "<value-b>"},
			),
		)
	})
})

var _ = Describe("func WithMessageComparator()", func() {
	It("configures how messages are compared", func() {
		handler := &IntegrationMessageHandler{
			ConfigureFunc: func(c dogma.IntegrationConfigurer) {
				c.Identity("<handler-name>", "<handler-key>")
				c.ConsumesCommandType(MessageC{})
				c.ProducesEventType(MessageE{})
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
				c.Identity("<app>", "<app-key>")
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
