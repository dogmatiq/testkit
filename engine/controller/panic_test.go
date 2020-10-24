package controller_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("func ConvertUnexpectedMessagePanic()", func() {
	config := configkit.FromProjection(
		&ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesEventType(MessageE{})
			},
		},
	)

	It("calls the function", func() {
		called := false

		ConvertUnexpectedMessagePanic(
			config,
			"<interface>",
			"<method>",
			MessageA1,
			func() {
				called = true
			},
		)

		Expect(called).To(BeTrue())
	})

	It("propagates panic values", func() {
		Expect(func() {
			ConvertUnexpectedMessagePanic(
				config,
				"<interface>",
				"<method>",
				MessageA1,
				func() {
					panic("<panic>")
				},
			)
		}).To(PanicWith("<panic>"))
	})

	It("converts UnexpectedMessage values", func() {
		Expect(func() {
			ConvertUnexpectedMessagePanic(
				config,
				"<interface>",
				"<method>",
				MessageA1,
				func() {
					panic(dogma.UnexpectedMessage)
				},
			)
		}).To(PanicWith(
			MatchAllFields(
				Fields{
					"Handler":   Equal(config),
					"Interface": Equal("<interface>"),
					"Method":    Equal("<method>"),
					"Message":   Equal(MessageA1),
					"PanicFunc": Not(BeEmpty()),
					"PanicFile": Not(BeEmpty()),
					"PanicLine": Not(BeZero()),
				},
			),
		))
	})
})
