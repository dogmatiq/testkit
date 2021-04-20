package testkit_test

import (
	"reflect"

	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit"
	"github.com/dogmatiq/testkit/internal/fixtures"
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
