package panicx_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/testkit/engine/panicx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("type UnexpectedMessage", func() {
	config := configkit.FromProjection(
		&ProjectionMessageHandler{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "<key>")
				c.ConsumesEventType(MessageE{})
			},
		},
	)

	Describe("func String()", func() {
		It("returns a description of the panic", func() {
			defer func() {
				r := recover()
				Expect(r).To(BeAssignableToTypeOf(UnexpectedMessage{}))

				x := r.(UnexpectedMessage)
				Expect(x.String()).To(Equal(
					"the '<name>' projection message handler did not expect <method>() to be called with a message of type fixtures.MessageA",
				))
			}()

			EnrichUnexpectedMessage(
				config,
				"<interface>",
				"<method>",
				MessageA1,
				func() {
					panic(dogma.UnexpectedMessage)
				},
			)
		})
	})
})

var _ = Describe("func EnrichUnexpectedMessage()", func() {
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

		EnrichUnexpectedMessage(
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
			EnrichUnexpectedMessage(
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
			EnrichUnexpectedMessage(
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
					"PanicLocation": MatchAllFields(
						Fields{
							"Func": Not(BeEmpty()),
							"File": HaveSuffix("/unexpectedmessage_test.go"),
							"Line": Equal(96),
						},
					),
				},
			),
		))
	})
})
