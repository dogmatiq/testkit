package controller_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("func ConvertUnexpectedMessagePanic()", func() {
	It("calls the function", func() {
		called := false

		ConvertUnexpectedMessagePanic(
			nil,
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
				nil,
				"<method>",
				MessageA1,
				func() {
					panic("<panic>")
				},
			)
		}).To(PanicWith("<panic>"))
	})

	It("converts UnexpectedMessage values", func() {
		config := configkit.FromProjection(
			&ProjectionMessageHandler{
				ConfigureFunc: func(c dogma.ProjectionConfigurer) {
					c.Identity("<name>", "<key>")
					c.ConsumesEventType(MessageE{})
				},
			},
		)

		Expect(func() {
			ConvertUnexpectedMessagePanic(
				config,
				"<method>",
				MessageA1,
				func() {
					panic(dogma.UnexpectedMessage)
				},
			)
		}).To(PanicWith(
			UnexpectedMessage{
				Handler: config,
				Method:  "<method>",
				Message: MessageA1,
			},
		))
	})
})
