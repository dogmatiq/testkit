package panicx_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/dogma/fixtures"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/panicx"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type UnexpectedMessage", func() {
	config := configkit.FromProjection(
		&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "a0eab8dd-db22-467a-87c2-c38138c582e8")
				c.Routes(
					dogma.HandlesEvent[MessageE](),
				)
			},
		},
	)

	g.Describe("func String()", func() {
		g.It("returns a description of the panic", func() {
			defer func() {
				r := recover()
				Expect(r).To(BeAssignableToTypeOf(UnexpectedMessage{}))

				x := r.(UnexpectedMessage)
				Expect(x.String()).To(Equal(
					"the '<name>' projection message handler did not expect *stubs.ProjectionMessageHandlerStub.<method>() to be called with a message of type fixtures.MessageA",
				))
			}()

			EnrichUnexpectedMessage(
				config,
				"<interface>",
				"<method>",
				config.Handler(),
				MessageA1,
				func() {
					panic(dogma.UnexpectedMessage)
				},
			)
		})
	})
})

var _ = g.Describe("func EnrichUnexpectedMessage()", func() {
	config := configkit.FromProjection(
		&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "b665eca3-936e-41e3-b9ab-c618cfa95ec2")
				c.Routes(
					dogma.HandlesEvent[MessageE](),
				)
			},
		},
	)

	g.It("calls the function", func() {
		called := false

		EnrichUnexpectedMessage(
			config,
			"<interface>",
			"<method>",
			config.Handler(),
			MessageA1,
			func() {
				called = true
			},
		)

		Expect(called).To(BeTrue())
	})

	g.It("propagates panic values", func() {
		Expect(func() {
			EnrichUnexpectedMessage(
				config,
				"<interface>",
				"<method>",
				config.Handler(),
				MessageA1,
				func() {
					panic("<panic>")
				},
			)
		}).To(PanicWith("<panic>"))
	})

	g.It("converts UnexpectedMessage values", func() {
		Expect(func() {
			EnrichUnexpectedMessage(
				config,
				"<interface>",
				"<method>",
				config.Handler(),
				MessageA1,
				doPanic,
			)
		}).To(PanicWith(
			MatchAllFields(
				Fields{
					"Handler":        Equal(config),
					"Interface":      Equal("<interface>"),
					"Method":         Equal("<method>"),
					"Implementation": Equal(config.Handler()),
					"Message":        Equal(MessageA1),
					"PanicLocation": MatchAllFields(
						Fields{
							"Func": Equal("github.com/dogmatiq/testkit/engine/internal/panicx_test.doPanic"),
							"File": HaveSuffix("/engine/internal/panicx/linenumber_test.go"),
							"Line": Equal(50),
						},
					),
				},
			),
		))
	})
})
