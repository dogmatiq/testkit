package panicx_test

import (
	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/dogma"
	. "github.com/dogmatiq/enginekit/enginetest/stubs"
	. "github.com/dogmatiq/testkit/engine/internal/panicx"
	g "github.com/onsi/ginkgo/v2"
	gm "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = g.Describe("type UnexpectedMessage", func() {
	config := configkit.FromProjection(
		&ProjectionMessageHandlerStub{
			ConfigureFunc: func(c dogma.ProjectionConfigurer) {
				c.Identity("<name>", "a0eab8dd-db22-467a-87c2-c38138c582e8")
				c.Routes(
					dogma.HandlesEvent[EventStub[TypeA]](),
				)
			},
		},
	)

	g.Describe("func String()", func() {
		g.It("returns a description of the panic", func() {
			defer func() {
				r := recover()
				gm.Expect(r).To(gm.BeAssignableToTypeOf(UnexpectedMessage{}))

				x := r.(UnexpectedMessage)
				gm.Expect(x.String()).To(gm.Equal(
					"the '<name>' projection message handler did not expect *stubs.ProjectionMessageHandlerStub.<method>() to be called with a message of type stubs.EventStub[TypeX]",
				))
			}()

			EnrichUnexpectedMessage(
				config,
				"<interface>",
				"<method>",
				config.Handler(),
				EventX1,
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
					dogma.HandlesEvent[EventStub[TypeA]](),
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
			EventX1,
			func() {
				called = true
			},
		)

		gm.Expect(called).To(gm.BeTrue())
	})

	g.It("propagates panic values", func() {
		gm.Expect(func() {
			EnrichUnexpectedMessage(
				config,
				"<interface>",
				"<method>",
				config.Handler(),
				EventX1,
				func() {
					panic("<panic>")
				},
			)
		}).To(gm.PanicWith("<panic>"))
	})

	g.It("converts UnexpectedMessage values", func() {
		gm.Expect(func() {
			EnrichUnexpectedMessage(
				config,
				"<interface>",
				"<method>",
				config.Handler(),
				EventX1,
				doPanic,
			)
		}).To(gm.PanicWith(
			MatchAllFields(
				Fields{
					"Handler":        gm.Equal(config),
					"Interface":      gm.Equal("<interface>"),
					"Method":         gm.Equal("<method>"),
					"Implementation": gm.Equal(config.Handler()),
					"Message":        gm.Equal(EventX1),
					"PanicLocation": MatchAllFields(
						Fields{
							"Func": gm.Equal("github.com/dogmatiq/testkit/engine/internal/panicx_test.doPanic"),
							"File": gm.HaveSuffix("/engine/internal/panicx/linenumber_test.go"),
							"Line": gm.Equal(50),
						},
					),
				},
			),
		))
	})
})
